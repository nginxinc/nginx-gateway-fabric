package provisioner

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/events"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/gatewayclass"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status2"
)

// eventHandler ensures each Gateway for the specific GatewayClass has a corresponding Deployment
// of NGF configured to use that specific Gateway.
//
// eventHandler implements events.Handler interface.
type eventHandler struct {
	gcName string
	store  *store

	// provisions maps NamespacedName of Gateway to its corresponding Deployment
	provisions map[types.NamespacedName]*v1.Deployment

	statusUpdater *status2.Updater
	k8sClient     client.Client

	staticModeDeploymentYAML []byte

	gatewayNextID int64
}

func newEventHandler(
	gcName string,
	statusUpdater *status2.Updater,
	k8sClient client.Client,
	staticModeDeploymentYAML []byte,
) *eventHandler {
	return &eventHandler{
		store:                    newStore(),
		provisions:               make(map[types.NamespacedName]*v1.Deployment),
		statusUpdater:            statusUpdater,
		gcName:                   gcName,
		k8sClient:                k8sClient,
		staticModeDeploymentYAML: staticModeDeploymentYAML,
		gatewayNextID:            1,
	}
}

func (h *eventHandler) setGatewayClassStatuses(ctx context.Context) {
	var reqs []status2.UpdateRequest

	var gcExists bool

	for nsname, gc := range h.store.gatewayClasses {
		// The order of conditions matters. Default conditions are added first so that any additional conditions will
		// override them, which is ensured by DeduplicateConditions.
		conds := conditions.NewDefaultGatewayClassConditions()

		if gc.Name == h.gcName {
			gcExists = true
		} else {
			conds = append(conds, conditions.NewGatewayClassConflict())
		}

		// We ignore the boolean return value here because the provisioner only sets status,
		// it does not generate config.
		supportedVersionConds, _ := gatewayclass.ValidateCRDVersions(h.store.crdMetadata)
		conds = append(conds, supportedVersionConds...)

		reqs = append(reqs, status2.UpdateRequest{
			NsName:       nsname,
			ResourceType: &gatewayv1.GatewayClass{},
			Setter: func(object client.Object) bool {
				gcs := gatewayv1.GatewayClassStatus{
					Conditions: status2.ConvertConditions(conditions.DeduplicateConditions(conds), gc.Generation, metav1.Now()),
				}

				if status2.ConditionsEqual(gc.Status.Conditions, gcs.Conditions) {
					return false
				}

				gc.Status = gcs

				return true
			},
		})
	}

	if !gcExists {
		panic(fmt.Errorf("GatewayClass %s must exist", h.gcName))
	}

	h.statusUpdater.Update(ctx, reqs...)
}

func (h *eventHandler) ensureDeploymentsMatchGateways(ctx context.Context, logger logr.Logger) {
	var gwsWithoutDeps, removedGwsWithDeps []types.NamespacedName

	for nsname, gw := range h.store.gateways {
		if string(gw.Spec.GatewayClassName) != h.gcName {
			continue
		}
		if _, exist := h.provisions[nsname]; exist {
			continue
		}

		gwsWithoutDeps = append(gwsWithoutDeps, nsname)
	}

	for nsname := range h.provisions {
		if _, exist := h.store.gateways[nsname]; exist {
			continue
		}

		removedGwsWithDeps = append(removedGwsWithDeps, nsname)
	}

	// Create new deployments

	for _, nsname := range gwsWithoutDeps {
		deployment, err := prepareDeployment(h.staticModeDeploymentYAML, h.generateDeploymentID(), nsname)
		if err != nil {
			panic(fmt.Errorf("failed to prepare deployment: %w", err))
		}

		if err = h.k8sClient.Create(ctx, deployment); err != nil {
			panic(fmt.Errorf("failed to create deployment: %w", err))
		}

		h.provisions[nsname] = deployment

		logger.Info(
			"Created deployment",
			"deployment", client.ObjectKeyFromObject(deployment),
			"gateway", nsname,
		)
	}

	// Remove unnecessary deployments

	for _, nsname := range removedGwsWithDeps {
		deployment := h.provisions[nsname]

		if err := h.k8sClient.Delete(ctx, deployment); err != nil {
			panic(fmt.Errorf("failed to delete deployment: %w", err))
		}

		delete(h.provisions, nsname)

		logger.Info(
			"Deleted deployment",
			"deployment", client.ObjectKeyFromObject(deployment),
			"gateway", nsname,
		)
	}
}

func (h *eventHandler) HandleEventBatch(ctx context.Context, logger logr.Logger, batch events.EventBatch) {
	h.store.update(batch)
	h.setGatewayClassStatuses(ctx)
	h.ensureDeploymentsMatchGateways(ctx, logger)
}

func (h *eventHandler) generateDeploymentID() string {
	// This approach will break if the provisioner is restarted, because the existing Gateways might get
	// IDs different from the previous replica of the provisioner.
	id := h.gatewayNextID
	h.gatewayNextID++

	return fmt.Sprintf("nginx-gateway-%d", id)
}
