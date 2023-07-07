package state

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/relationship"
)

// Updater updates the cluster state.
type Updater interface {
	Upsert(obj client.Object)
	Delete(objType client.Object, nsname types.NamespacedName)
}

// objectStore is a store of client.Object
type objectStore interface {
	get(nsname types.NamespacedName) (client.Object, bool)
	upsert(obj client.Object)
	delete(nsname types.NamespacedName)
}

// objectStoreMapAdapter wraps maps of types.NamespacedName to Kubernetes resources
// (e.g. map[types.NamespacedName]*v1beta1.Gateway) so that they can be used through objectStore interface.
type objectStoreMapAdapter[T client.Object] struct {
	objects map[types.NamespacedName]T
}

func newObjectStoreMapAdapter[T client.Object](objects map[types.NamespacedName]T) *objectStoreMapAdapter[T] {
	return &objectStoreMapAdapter[T]{
		objects: objects,
	}
}

func (m *objectStoreMapAdapter[T]) get(nsname types.NamespacedName) (client.Object, bool) {
	obj, exist := m.objects[nsname]
	return obj, exist
}

func (m *objectStoreMapAdapter[T]) upsert(obj client.Object) {
	t, ok := obj.(T)
	if !ok {
		panic(fmt.Errorf("obj type mismatch: got %T, expected %T", obj, t))
	}
	m.objects[client.ObjectKeyFromObject(obj)] = t
}

func (m *objectStoreMapAdapter[T]) delete(nsname types.NamespacedName) {
	delete(m.objects, nsname)
}

type gvkList []schema.GroupVersionKind

func (list gvkList) contains(gvk schema.GroupVersionKind) bool {
	for _, g := range list {
		if gvk == g {
			return true
		}
	}

	return false
}

type multiObjectStore struct {
	stores     map[schema.GroupVersionKind]objectStore
	extractGVK extractGVKFunc
}

func newMultiObjectStore(
	stores map[schema.GroupVersionKind]objectStore,
	extractGVK extractGVKFunc,
) *multiObjectStore {
	return &multiObjectStore{
		stores:     stores,
		extractGVK: extractGVK,
	}
}

func (m *multiObjectStore) mustFindStoreForObj(obj client.Object) objectStore {
	objGVK := m.extractGVK(obj)

	store, exist := m.stores[objGVK]
	if !exist {
		panic(fmt.Errorf("object store for %T %v not found", obj, client.ObjectKeyFromObject(obj)))
	}

	return store
}

func (m *multiObjectStore) get(objType client.Object, nsname types.NamespacedName) (client.Object, bool) {
	return m.mustFindStoreForObj(objType).get(nsname)
}

func (m *multiObjectStore) upsert(obj client.Object) {
	m.mustFindStoreForObj(obj).upsert(obj)
}

func (m *multiObjectStore) delete(objType client.Object, nsname types.NamespacedName) {
	m.mustFindStoreForObj(objType).delete(nsname)
}

type changeTrackingUpdaterObjectTypeCfg struct {
	// store holds the objects of the gvk. If the store is nil, the objects of the gvk are not persisted.
	store objectStore
	gvk   schema.GroupVersionKind
	// trackUpsertDelete indicates whether an upsert or delete of an object with the gvk results into a change to
	// the changeTrackingUpdater's store. Note that for an upsert, the generation of a new object must be different
	// from the generation of the previous version, otherwise such an upsert is not considered a change.
	trackUpsertDelete bool
}

// triggerStateChangeFunc triggers a change to the changeTrackingUpdater's store for the given object.
type triggerStateChangeFunc func(objType client.Object, nsname types.NamespacedName) bool

// changeTrackingUpdater is an Updater that tracks changes to the cluster state in the multiObjectStore.
//
// It only works with objects with the GVKs registered in changeTrackingUpdaterObjectTypeCfg. Otherwise, it panics.
//
// A change is tracked when:
// - An object with a GVK with a non-nil store and trackUpsertDelete set to 'true' is upserted or deleted, provided
// that its generation changed.
// - An object is upserted or deleted, and it is related to another object, based on the decision by
// the relationship capturer.
// - An object is upserted or deleted and triggerStateChange returns true for the object.
type changeTrackingUpdater struct {
	store              *multiObjectStore
	capturer           relationship.Capturer
	triggerStateChange triggerStateChangeFunc

	extractGVK              extractGVKFunc
	supportedGVKs           gvkList
	trackedUpsertDeleteGVKs gvkList
	persistedGVKs           gvkList

	changed bool
}

func newChangeTrackingUpdater(
	capturer relationship.Capturer,
	triggerStateChange triggerStateChangeFunc,
	extractGVK extractGVKFunc,
	objectTypeCfgs []changeTrackingUpdaterObjectTypeCfg,
) *changeTrackingUpdater {
	var (
		supportedGVKs           gvkList
		trackedUpsertDeleteGVKs gvkList
		persistedGVKs           gvkList

		stores = make(map[schema.GroupVersionKind]objectStore)
	)

	for _, cfg := range objectTypeCfgs {
		supportedGVKs = append(supportedGVKs, cfg.gvk)

		if cfg.trackUpsertDelete {
			trackedUpsertDeleteGVKs = append(trackedUpsertDeleteGVKs, cfg.gvk)
		}

		if cfg.store != nil {
			persistedGVKs = append(persistedGVKs, cfg.gvk)
			stores[cfg.gvk] = cfg.store
		}
	}

	return &changeTrackingUpdater{
		store:                   newMultiObjectStore(stores, extractGVK),
		extractGVK:              extractGVK,
		supportedGVKs:           supportedGVKs,
		trackedUpsertDeleteGVKs: trackedUpsertDeleteGVKs,
		persistedGVKs:           persistedGVKs,
		capturer:                capturer,
		triggerStateChange:      triggerStateChange,
	}
}

func (s *changeTrackingUpdater) assertSupportedGVK(gvk schema.GroupVersionKind) {
	if !s.supportedGVKs.contains(gvk) {
		panic(fmt.Errorf("unsupported GVK %v", gvk))
	}
}

func (s *changeTrackingUpdater) upsert(obj client.Object) (changed bool) {
	if !s.persistedGVKs.contains(s.extractGVK(obj)) {
		return false
	}

	oldObj, exist := s.store.get(obj, client.ObjectKeyFromObject(obj))
	s.store.upsert(obj)

	if !s.trackedUpsertDeleteGVKs.contains(s.extractGVK(obj)) {
		return false
	}

	return !exist || obj.GetGeneration() != oldObj.GetGeneration()
}

func (s *changeTrackingUpdater) Upsert(obj client.Object) {
	s.assertSupportedGVK(s.extractGVK(obj))

	changingUpsert := s.upsert(obj)
	relationshipExisted := s.capturer.Exists(obj, client.ObjectKeyFromObject(obj))

	s.capturer.Capture(obj)

	relationshipExists := s.capturer.Exists(obj, client.ObjectKeyFromObject(obj))

	forceChanged := s.triggerStateChange(obj, client.ObjectKeyFromObject(obj))

	// FIXME(pleshakov): Check generation in all cases to minimize the number of Graph regeneration.
	// s.changed can be true even if the generation of the object did not change, because
	// capturer and triggerStateChange don't take the generation into account.
	// See https://github.com/nginxinc/nginx-kubernetes-gateway/issues/825

	s.changed = s.changed || changingUpsert || relationshipExisted || relationshipExists || forceChanged
}

func (s *changeTrackingUpdater) delete(objType client.Object, nsname types.NamespacedName) (changed bool) {
	objTypeGVK := s.extractGVK(objType)

	if !s.persistedGVKs.contains(objTypeGVK) {
		return false
	}

	_, exist := s.store.get(objType, nsname)
	if !exist {
		return false
	}
	s.store.delete(objType, nsname)

	return s.trackedUpsertDeleteGVKs.contains(objTypeGVK)
}

func (s *changeTrackingUpdater) Delete(objType client.Object, nsname types.NamespacedName) {
	s.assertSupportedGVK(s.extractGVK(objType))

	changingDelete := s.delete(objType, nsname)

	forceChanged := s.triggerStateChange(objType, nsname)

	s.changed = s.changed || changingDelete || s.capturer.Exists(objType, nsname) || forceChanged

	s.capturer.Remove(objType, nsname)
}

// getAndResetChangedStatus returns true if the previous updates (Upserts/Deletes) require an update of
// the configuration of the data plane. It also resets the changed status to false.
func (s *changeTrackingUpdater) getAndResetChangedStatus() bool {
	changed := s.changed
	s.changed = false
	return changed
}

type upsertValidatorFunc func(obj client.Object) error

// validatingUpsertUpdater is an Updater that validates an object before upserting it.
// If the validation fails, it deletes the object and records an event with the validation error.
type validatingUpsertUpdater struct {
	updater       Updater
	eventRecorder record.EventRecorder
	validator     upsertValidatorFunc
}

func newValidatingUpsertUpdater(
	updater Updater,
	eventRecorder record.EventRecorder,
	validator upsertValidatorFunc,
) *validatingUpsertUpdater {
	return &validatingUpsertUpdater{
		updater:       updater,
		eventRecorder: eventRecorder,
		validator:     validator,
	}
}

func (u *validatingUpsertUpdater) Upsert(obj client.Object) {
	if err := u.validator(obj); err != nil {
		u.updater.Delete(obj, client.ObjectKeyFromObject(obj))

		u.eventRecorder.Eventf(
			obj,
			apiv1.EventTypeWarning,
			"Rejected",
			"%s; NKG will delete any existing NGINX configuration that corresponds to the resource",
			err.Error(),
		)

		return
	}

	u.updater.Upsert(obj)
}

func (u *validatingUpsertUpdater) Delete(objType client.Object, nsname types.NamespacedName) {
	u.updater.Delete(objType, nsname)
}
