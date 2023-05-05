package main

import (
	"bytes"
	"context"
	"text/template"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
)

type eventHandler struct {
	gatewayClasses map[types.NamespacedName]*v1beta1.GatewayClass
	gateways       map[types.NamespacedName]*v1beta1.Gateway

	provisions map[types.NamespacedName]*v1.Deployment

	statusUpdater status.Updater

	gcName string

	k8sClient client.Client

	logger logr.Logger
}

func newEventHandler(
	gcName string,
	statusUpdater status.Updater,
	k8sClient client.Client,
	logger logr.Logger,
) *eventHandler {
	return &eventHandler{
		gatewayClasses: make(map[types.NamespacedName]*v1beta1.GatewayClass),
		gateways:       make(map[types.NamespacedName]*v1beta1.Gateway),
		provisions:     make(map[types.NamespacedName]*v1.Deployment),
		statusUpdater:  statusUpdater,
		gcName:         gcName,
		k8sClient:      k8sClient,
		logger:         logger,
	}
}

func (h *eventHandler) HandleEventBatch(ctx context.Context, batch events.EventBatch) {
	// update caches
	for _, event := range batch {
		switch e := event.(type) {
		case *events.UpsertEvent:
			switch obj := e.Resource.(type) {
			case *v1beta1.GatewayClass:
				h.gatewayClasses[client.ObjectKeyFromObject(obj)] = obj
			case *v1beta1.Gateway:
				h.gateways[client.ObjectKeyFromObject(obj)] = obj
			default:
				panic("unknown object type")
			}
		case *events.DeleteEvent:
			switch e.Type.(type) {
			case *v1beta1.GatewayClass:
				delete(h.gatewayClasses, e.NamespacedName)
			case *v1beta1.Gateway:
				delete(h.gateways, e.NamespacedName)
			default:
				panic("unknown object type")
			}
		default:
			panic("unknown event type")
		}
	}

	// set Accepted True for our GatewayClass

	gc, exist := h.gatewayClasses[types.NamespacedName{Name: h.gcName}]
	if !exist {
		panic("gateway class not found")
	}

	statuses := state.Statuses{
		GatewayClassStatus: &state.GatewayClassStatus{
			Conditions:         conditions.NewDefaultGatewayClassConditions(),
			ObservedGeneration: gc.Generation,
		},
	}

	// process Gateways

	var toCreate, toRemove []types.NamespacedName

	for nsname, gw := range h.gateways {
		if string(gw.Spec.GatewayClassName) != h.gcName {
			continue
		}

		_, exist := h.provisions[nsname]
		if exist {
			continue
		}

		toCreate = append(toCreate, nsname)
	}

	for nsname := range h.provisions {
		_, exist := h.gateways[nsname]
		if exist {
			continue
		}

		toRemove = append(toRemove, nsname)
	}

	// create new deployments

	for _, nsname := range toCreate {
		// create deployment

		deployment := prepareDeployment(nsname.Name)

		err := h.k8sClient.Create(ctx, deployment)
		if err != nil {
			panic(err)
		}

		h.provisions[nsname] = deployment

		h.logger.Info("created deployment", "deployment", client.ObjectKeyFromObject(deployment))
	}

	// remove unnecessary deployments

	for _, nsname := range toRemove {
		deployment := h.provisions[nsname]

		err := h.k8sClient.Delete(ctx, deployment)
		if err != nil {
			panic(err)
		}

		delete(h.provisions, nsname)

		h.logger.Info("deleted deployment", "deployment", client.ObjectKeyFromObject(deployment))
	}

	// update statuses

	h.statusUpdater.Update(ctx, statuses)
}

const deploymentTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Name }} 
  namespace: nginx-gateway
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{ .Name }} 
  template:
    metadata:
      labels:
        app: {{ .Name }} 
    spec:
      shareProcessNamespace: true
      serviceAccountName: nginx-gateway
      volumes:
      - name: nginx-config
        emptyDir: { }
      - name: var-lib-nginx
        emptyDir: { }
      - name: njs-modules
        configMap:
          name: njs-modules
      initContainers:
      - image: busybox:1.34 # FIXME(pleshakov): use gateway container to init the Config with proper main config
        name: nginx-config-initializer
        command: [ 'sh', '-c', 'echo "load_module /usr/lib/nginx/modules/ngx_http_js_module.so; events {}  pid /etc/nginx/nginx.pid; error_log stderr debug; http { include /etc/nginx/conf.d/*.conf; js_import /usr/lib/nginx/modules/njs/httpmatches.js; }" > /etc/nginx/nginx.conf && (rm -r /etc/nginx/conf.d /etc/nginx/secrets; mkdir /etc/nginx/conf.d /etc/nginx/secrets && chown 1001:0 /etc/nginx/conf.d /etc/nginx/secrets)' ]
        volumeMounts:
        - name: nginx-config
          mountPath: /etc/nginx
      containers:
      - image: nginx-kubernetes-gateway:edge
        imagePullPolicy: Never 
        name: nginx-gateway
        volumeMounts:
        - name: nginx-config
          mountPath: /etc/nginx
        securityContext:
          runAsUser: 1001
          # FIXME(pleshakov) - figure out which capabilities are required
          # dropping ALL and adding only CAP_KILL doesn't work
          # Note: CAP_KILL is needed for sending HUP signal to NGINX main process
        args:
        - --gateway-ctlr-name=k8s-gateway.nginx.org/nginx-gateway-controller
        - --gatewayclass=nginx
      - image: nginx:1.23
        imagePullPolicy: IfNotPresent
        name: nginx
        ports:
        - name: http
          containerPort: 80
        - name: https
          containerPort: 443
        volumeMounts:
        - name: nginx-config
          mountPath: /etc/nginx
        - name: var-lib-nginx
          mountPath: /var/lib/nginx
        - name: njs-modules
          mountPath: /usr/lib/nginx/modules/njs
`

func prepareDeployment(gwName string) *v1.Deployment {
	t := template.Must(template.New("deployment").Parse(deploymentTemplate))

	var buf bytes.Buffer
	err := t.Execute(&buf, struct{ Name string }{Name: gwName})
	if err != nil {
		panic(err)
	}

	deployment := &v1.Deployment{}
	err = yaml.Unmarshal(buf.Bytes(), deployment)
	if err != nil {
		panic(err)
	}

	return deployment
}
