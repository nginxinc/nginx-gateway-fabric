package configuration

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

type Configuration struct {
	gcName string

	gc *v1alpha2.GatewayClass
	gw *v1alpha2.Gateway
	httpRoutes map[string]*v1alpha2.HTTPRoute

	httpListeners map[string]*httpListener

	changeCh chan Change
	notificationCh chan Notification
}

type httpListener struct {
	hosts map[string]*Host
	httpRoutes map[string]string
}

type Host struct {
	Value string
	Routes []*Route
}

type PathRoute struct {
	Path string
	// sorted based on the creation timestamp and namespace/name of the source
	// sorting must be stable to preserve the order in HTTPRoute
	Routes []Route
}

type Route struct {
	Rule *v1alpha2.HTTPRouteRule
	Source *v1alpha2.HTTPRoute
}

type Operation int

const (
	Delete Operation = iota
	Upsert
)

type Change struct {
	Op Operation
	Host *Host
}

type Notification struct {
	Object runtime.Object
	Reason string
	Message string
}

type Event struct {
	change *Change
}

func NewConfiguration(gcName string) *Configuration {
	return &Configuration{
		gcName: gcName,
	}
}

func (c *Configuration) GetChangeCh() <-chan Change {
	return c.changeCh
}

func (c *Configuration) GetNotificationCh() <-chan Notification {
	return c.notificationCh
}

func (c *Configuration) UpsertGatewayClass(gc *v1alpha2.GatewayClass) {
	// validate

	// if not valid
	// create changes (to remove any existing configuration)
	// create notification to update status
	// return

	// create notification to update status
}

func (c *Configuration) DeleteGatewayClass() {
	// create changes (to remove any existing configuration)
}

func (c *Configuration) UpsertGateway(gc *v1alpha2.Gateway) {
	// if no gateway class
	// reject
	// return

	// validate

	// if not valid
	// remove the httpListener
	// create changes based on that
	// create notifications
	// return

	// rebuild httpListener
	// create changes based on that
	// create notifications
}

func (c *Configuration) DeleteGateway() {
	// create changes based on that
	// create notifications
}

func (c *Configuration) UpsertHTTPRoute(httpRoute *v1alpha2.HTTPRoute) {
	// validate

	// if not valid
	// remove from the c.httpRoutes map
	// rebuild hosts of the httpListener
	// create changes
	// put changes in the ChangeCH
	// return

	key := getResourceKey(&httpRoute.ObjectMeta)

	c.httpRoutes[key] = httpRoute

	// if http listener is not configured, report an error (notification)
	// if c.httpListener == nil {
	// 	notif := Notification{
	// 		Object:  httpRoute,
	// 		Reason:  "Ignored",
	// 		Message: "Listener doesn't exist",
	// 	}
	//
	// 	c.notificationCh <- notif
	// 	return
	// }

	// rebuild hosts in the httpListener hosts
	// compare old and new ones
	// build changes based on that
	// put changes in the ChangeCh
}

func (c *Configuration) DeleteHTTPRoute(key string) {
	// delete from the map
	// update the hosts of the HTTPListener
	// determinate changes
	// put changes in the ChangeCh
}

func getResourceKey(meta *metav1.ObjectMeta) string {
	return fmt.Sprintf("%s/%s", meta.Namespace, meta.Name)
}