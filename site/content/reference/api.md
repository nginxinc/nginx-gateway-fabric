---
title: "API Reference"
description: "NGINX Gateway API Reference"
weight: 100
toc: false
---
<p>Packages:</p>
<ul>
<li>
<a href="#gateway.nginx.org%2fv1alpha1">gateway.nginx.org/v1alpha1</a>
</li>
</ul>
<h2 id="gateway.nginx.org/v1alpha1">gateway.nginx.org/v1alpha1</h2>
<div>
<p>Package v1alpha1 contains API Schema definitions for the
gateway.nginx.org API group.</p>
</div>
<ul></ul>
<h3 id="gateway.nginx.org/v1alpha1.ClientBody">ClientBody
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.ClientSettingsPolicySpec">ClientSettingsPolicySpec</a>)
</p>
<div>
<p>ClientBody contains the settings for the client request body.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>maxSize</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.Size">
Size
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>MaxSize sets the maximum allowed size of the client request body.
If the size in a request exceeds the configured value,
the 413 (Request Entity Too Large) error is returned to the client.
Setting size to 0 disables checking of client request body size.
Default: <a href="https://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size">https://nginx.org/en/docs/http/ngx_http_core_module.html#client_max_body_size</a>.</p>
</td>
</tr>
<tr>
<td>
<code>timeout</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Timeout defines a timeout for reading client request body. The timeout is set only for a period between
two successive read operations, not for the transmission of the whole request body.
If a client does not transmit anything within this time, the request is terminated with the
408 (Request Time-out) error.
Default: <a href="https://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_timeout">https://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_timeout</a>.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.ClientKeepAlive">ClientKeepAlive
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.ClientSettingsPolicySpec">ClientSettingsPolicySpec</a>)
</p>
<div>
<p>ClientKeepAlive defines the keep-alive settings for clients.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>requests</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Requests sets the maximum number of requests that can be served through one keep-alive connection.
After the maximum number of requests are made, the connection is closed. Closing connections periodically
is necessary to free per-connection memory allocations. Therefore, using too high maximum number of requests
is not recommended as it can lead to excessive memory usage.
Default: <a href="https://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_requests">https://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_requests</a>.</p>
</td>
</tr>
<tr>
<td>
<code>time</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Time defines the maximum time during which requests can be processed through one keep-alive connection.
After this time is reached, the connection is closed following the subsequent request processing.
Default: <a href="https://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_time">https://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_time</a>.</p>
</td>
</tr>
<tr>
<td>
<code>timeout</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.ClientKeepAliveTimeout">
ClientKeepAliveTimeout
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Timeout defines the keep-alive timeouts for clients.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.ClientKeepAliveTimeout">ClientKeepAliveTimeout
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.ClientKeepAlive">ClientKeepAlive</a>)
</p>
<div>
<p>ClientKeepAliveTimeout defines the timeouts related to keep-alive client connections.
Default: Default: <a href="https://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout">https://nginx.org/en/docs/http/ngx_http_core_module.html#keepalive_timeout</a>.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>server</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Server sets the timeout during which a keep-alive client connection will stay open on the server side.
Setting this value to 0 disables keep-alive client connections.</p>
</td>
</tr>
<tr>
<td>
<code>header</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Header sets the timeout in the &ldquo;Keep-Alive: timeout=time&rdquo; response header field.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.ClientSettingsPolicy">ClientSettingsPolicy
</h3>
<div>
<p>ClientSettingsPolicy is an Inherited Attached Policy. It provides a way to configure the behavior of the connection
between the client and NGINX Gateway Fabric.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.ClientSettingsPolicySpec">
ClientSettingsPolicySpec
</a>
</em>
</td>
<td>
<p>Spec defines the desired state of the ClientSettingsPolicy.</p>
<br/>
<br/>
<table class="table table-bordered table-striped">
<tr>
<td>
<code>body</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.ClientBody">
ClientBody
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Body defines the client request body settings.</p>
</td>
</tr>
<tr>
<td>
<code>keepAlive</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.ClientKeepAlive">
ClientKeepAlive
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>KeepAlive defines the keep-alive settings.</p>
</td>
</tr>
<tr>
<td>
<code>targetRef</code><br/>
<em>
<a href="https://pkg.go.dev/sigs.k8s.io/gateway-api/apis/v1alpha2#LocalPolicyTargetReference">
sigs.k8s.io/gateway-api/apis/v1alpha2.LocalPolicyTargetReference
</a>
</em>
</td>
<td>
<p>TargetRef identifies an API object to apply the policy to.
Object must be in the same namespace as the policy.
Support: Gateway, HTTPRoute, GRPCRoute.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="https://pkg.go.dev/sigs.k8s.io/gateway-api/apis/v1alpha2#PolicyStatus">
sigs.k8s.io/gateway-api/apis/v1alpha2.PolicyStatus
</a>
</em>
</td>
<td>
<p>Status defines the state of the ClientSettingsPolicy.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.ClientSettingsPolicySpec">ClientSettingsPolicySpec
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.ClientSettingsPolicy">ClientSettingsPolicy</a>)
</p>
<div>
<p>ClientSettingsPolicySpec defines the desired state of ClientSettingsPolicy.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>body</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.ClientBody">
ClientBody
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Body defines the client request body settings.</p>
</td>
</tr>
<tr>
<td>
<code>keepAlive</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.ClientKeepAlive">
ClientKeepAlive
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>KeepAlive defines the keep-alive settings.</p>
</td>
</tr>
<tr>
<td>
<code>targetRef</code><br/>
<em>
<a href="https://pkg.go.dev/sigs.k8s.io/gateway-api/apis/v1alpha2#LocalPolicyTargetReference">
sigs.k8s.io/gateway-api/apis/v1alpha2.LocalPolicyTargetReference
</a>
</em>
</td>
<td>
<p>TargetRef identifies an API object to apply the policy to.
Object must be in the same namespace as the policy.
Support: Gateway, HTTPRoute, GRPCRoute.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.ControllerLogLevel">ControllerLogLevel
(<code>string</code> alias)</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.Logging">Logging</a>)
</p>
<div>
<p>ControllerLogLevel type defines the logging level for the control plane.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;debug&#34;</p></td>
<td><p>ControllerLogLevelDebug is the debug level for control plane logging.</p>
</td>
</tr><tr><td><p>&#34;error&#34;</p></td>
<td><p>ControllerLogLevelError is the error level for control plane logging.</p>
</td>
</tr><tr><td><p>&#34;info&#34;</p></td>
<td><p>ControllerLogLevelInfo is the info level for control plane logging.</p>
</td>
</tr></tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.Duration">Duration
(<code>string</code> alias)</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.ClientBody">ClientBody</a>, <a href="#gateway.nginx.org/v1alpha1.ClientKeepAlive">ClientKeepAlive</a>, <a href="#gateway.nginx.org/v1alpha1.ClientKeepAliveTimeout">ClientKeepAliveTimeout</a>, <a href="#gateway.nginx.org/v1alpha1.TelemetryExporter">TelemetryExporter</a>)
</p>
<div>
<p>Duration is a string value representing a duration in time.
Duration can be specified in milliseconds (ms) or seconds (s) A value without a suffix is seconds.
Examples: 120s, 50ms.</p>
</div>
<h3 id="gateway.nginx.org/v1alpha1.Logging">Logging
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.NginxGatewaySpec">NginxGatewaySpec</a>)
</p>
<div>
<p>Logging defines logging related settings for the control plane.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>level</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.ControllerLogLevel">
ControllerLogLevel
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Level defines the logging level.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.NginxGateway">NginxGateway
</h3>
<div>
<p>NginxGateway represents the dynamic configuration for an NGINX Gateway Fabric control plane.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.NginxGatewaySpec">
NginxGatewaySpec
</a>
</em>
</td>
<td>
<p>NginxGatewaySpec defines the desired state of the NginxGateway.</p>
<br/>
<br/>
<table class="table table-bordered table-striped">
<tr>
<td>
<code>logging</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.Logging">
Logging
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Logging defines logging related settings for the control plane.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.NginxGatewayStatus">
NginxGatewayStatus
</a>
</em>
</td>
<td>
<p>NginxGatewayStatus defines the state of the NginxGateway.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.NginxGatewayConditionReason">NginxGatewayConditionReason
(<code>string</code> alias)</h3>
<div>
<p>NginxGatewayConditionReason defines the set of reasons that explain why a
particular NginxGateway condition type has been raised.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Invalid&#34;</p></td>
<td><p>NginxGatewayReasonInvalid is a reason that is used with the &ldquo;Valid&rdquo; condition when the condition is False.</p>
</td>
</tr><tr><td><p>&#34;Valid&#34;</p></td>
<td><p>NginxGatewayReasonValid is a reason that is used with the &ldquo;Valid&rdquo; condition when the condition is True.</p>
</td>
</tr></tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.NginxGatewayConditionType">NginxGatewayConditionType
(<code>string</code> alias)</h3>
<div>
<p>NginxGatewayConditionType is a type of condition associated with an
NginxGateway. This type should be used with the NginxGatewayStatus.Conditions field.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Valid&#34;</p></td>
<td><p>NginxGatewayConditionValid is a condition that is true when the NginxGateway
configuration is syntactically and semantically valid.</p>
</td>
</tr></tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.NginxGatewaySpec">NginxGatewaySpec
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.NginxGateway">NginxGateway</a>)
</p>
<div>
<p>NginxGatewaySpec defines the desired state of the NginxGateway.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>logging</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.Logging">
Logging
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Logging defines logging related settings for the control plane.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.NginxGatewayStatus">NginxGatewayStatus
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.NginxGateway">NginxGateway</a>)
</p>
<div>
<p>NginxGatewayStatus defines the state of the NginxGateway.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#condition-v1-meta">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.NginxProxy">NginxProxy
</h3>
<div>
<p>NginxProxy is a configuration object that is attached to a GatewayClass parametersRef. It provides a way
to configure global settings for all Gateways defined from the GatewayClass.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.NginxProxySpec">
NginxProxySpec
</a>
</em>
</td>
<td>
<p>Spec defines the desired state of the NginxProxy.</p>
<br/>
<br/>
<table class="table table-bordered table-striped">
<tr>
<td>
<code>telemetry</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.Telemetry">
Telemetry
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Telemetry specifies the OpenTelemetry configuration.</p>
</td>
</tr>
<tr>
<td>
<code>disableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>DisableHTTP2 defines if http2 should be disabled for all servers.
Default is false, meaning http2 will be enabled for all servers.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.NginxProxySpec">NginxProxySpec
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.NginxProxy">NginxProxy</a>)
</p>
<div>
<p>NginxProxySpec defines the desired state of the NginxProxy.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>telemetry</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.Telemetry">
Telemetry
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Telemetry specifies the OpenTelemetry configuration.</p>
</td>
</tr>
<tr>
<td>
<code>disableHTTP2</code><br/>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>DisableHTTP2 defines if http2 should be disabled for all servers.
Default is false, meaning http2 will be enabled for all servers.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.ObservabilityPolicy">ObservabilityPolicy
</h3>
<div>
<p>ObservabilityPolicy is a Direct Attached Policy. It provides a way to configure observability settings for
the NGINX Gateway Fabric data plane. Used in conjunction with the NginxProxy CRD that is attached to the
GatewayClass parametersRef.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.24/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.ObservabilityPolicySpec">
ObservabilityPolicySpec
</a>
</em>
</td>
<td>
<p>Spec defines the desired state of the ObservabilityPolicy.</p>
<br/>
<br/>
<table class="table table-bordered table-striped">
<tr>
<td>
<code>tracing</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.Tracing">
Tracing
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Tracing allows for enabling and configuring tracing.</p>
</td>
</tr>
<tr>
<td>
<code>targetRefs</code><br/>
<em>
<a href="https://pkg.go.dev/sigs.k8s.io/gateway-api/apis/v1alpha2#LocalPolicyTargetReference">
[]sigs.k8s.io/gateway-api/apis/v1alpha2.LocalPolicyTargetReference
</a>
</em>
</td>
<td>
<p>TargetRefs identifies the API object(s) to apply the policy to.
Objects must be in the same namespace as the policy.
Support: HTTPRoute</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="https://pkg.go.dev/sigs.k8s.io/gateway-api/apis/v1alpha2#PolicyStatus">
sigs.k8s.io/gateway-api/apis/v1alpha2.PolicyStatus
</a>
</em>
</td>
<td>
<p>Status defines the state of the ObservabilityPolicy.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.ObservabilityPolicySpec">ObservabilityPolicySpec
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.ObservabilityPolicy">ObservabilityPolicy</a>)
</p>
<div>
<p>ObservabilityPolicySpec defines the desired state of the ObservabilityPolicy.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>tracing</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.Tracing">
Tracing
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Tracing allows for enabling and configuring tracing.</p>
</td>
</tr>
<tr>
<td>
<code>targetRefs</code><br/>
<em>
<a href="https://pkg.go.dev/sigs.k8s.io/gateway-api/apis/v1alpha2#LocalPolicyTargetReference">
[]sigs.k8s.io/gateway-api/apis/v1alpha2.LocalPolicyTargetReference
</a>
</em>
</td>
<td>
<p>TargetRefs identifies the API object(s) to apply the policy to.
Objects must be in the same namespace as the policy.
Support: HTTPRoute</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.Size">Size
(<code>string</code> alias)</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.ClientBody">ClientBody</a>)
</p>
<div>
<p>Size is a string value representing a size. Size can be specified in bytes, kilobytes (k), megabytes (m),
or gigabytes (g).
Examples: 1024, 8k, 1m.</p>
</div>
<h3 id="gateway.nginx.org/v1alpha1.SpanAttribute">SpanAttribute
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.Telemetry">Telemetry</a>, <a href="#gateway.nginx.org/v1alpha1.Tracing">Tracing</a>)
</p>
<div>
<p>SpanAttribute is a key value pair to be added to a tracing span.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<p>Key is the key for a span attribute.
Format: must have all &lsquo;&ldquo;&rsquo; escaped and must not contain any &lsquo;$&rsquo; or end with an unescaped &lsquo;\&rsquo;</p>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
string
</em>
</td>
<td>
<p>Value is the value for a span attribute.
Format: must have all &lsquo;&ldquo;&rsquo; escaped and must not contain any &lsquo;$&rsquo; or end with an unescaped &lsquo;\&rsquo;</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.Telemetry">Telemetry
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.NginxProxySpec">NginxProxySpec</a>)
</p>
<div>
<p>Telemetry specifies the OpenTelemetry configuration.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>exporter</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.TelemetryExporter">
TelemetryExporter
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Exporter specifies OpenTelemetry export parameters.</p>
</td>
</tr>
<tr>
<td>
<code>serviceName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ServiceName is the &ldquo;service.name&rdquo; attribute of the OpenTelemetry resource.
Default is &lsquo;ngf:<gateway-namespace>:<gateway-name>&rsquo;. If a value is provided by the user,
then the default becomes a prefix to that value.</p>
</td>
</tr>
<tr>
<td>
<code>spanAttributes</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.SpanAttribute">
[]SpanAttribute
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SpanAttributes are custom key/value attributes that are added to each span.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.TelemetryExporter">TelemetryExporter
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.Telemetry">Telemetry</a>)
</p>
<div>
<p>TelemetryExporter specifies OpenTelemetry export parameters.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>interval</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.Duration">
Duration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Interval is the maximum interval between two exports.
Default: <a href="https://nginx.org/en/docs/ngx_otel_module.html#otel_exporter">https://nginx.org/en/docs/ngx_otel_module.html#otel_exporter</a></p>
</td>
</tr>
<tr>
<td>
<code>batchSize</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>BatchSize is the maximum number of spans to be sent in one batch per worker.
Default: <a href="https://nginx.org/en/docs/ngx_otel_module.html#otel_exporter">https://nginx.org/en/docs/ngx_otel_module.html#otel_exporter</a></p>
</td>
</tr>
<tr>
<td>
<code>batchCount</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>BatchCount is the number of pending batches per worker, spans exceeding the limit are dropped.
Default: <a href="https://nginx.org/en/docs/ngx_otel_module.html#otel_exporter">https://nginx.org/en/docs/ngx_otel_module.html#otel_exporter</a></p>
</td>
</tr>
<tr>
<td>
<code>endpoint</code><br/>
<em>
string
</em>
</td>
<td>
<p>Endpoint is the address of OTLP/gRPC endpoint that will accept telemetry data.
Format: alphanumeric hostname with optional http scheme and optional port.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.TraceContext">TraceContext
(<code>string</code> alias)</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.Tracing">Tracing</a>)
</p>
<div>
<p>TraceContext specifies how to propagate traceparent/tracestate headers.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;extract&#34;</p></td>
<td><p>TraceContextExtract uses an existing trace context from the request, so that the identifiers
of a trace and the parent span are inherited from the incoming request.</p>
</td>
</tr><tr><td><p>&#34;ignore&#34;</p></td>
<td><p>TraceContextIgnore skips context headers processing.</p>
</td>
</tr><tr><td><p>&#34;inject&#34;</p></td>
<td><p>TraceContextInject adds a new context to the request, overwriting existing headers, if any.</p>
</td>
</tr><tr><td><p>&#34;propagate&#34;</p></td>
<td><p>TraceContextPropagate updates the existing context (combines extract and inject).</p>
</td>
</tr></tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.TraceStrategy">TraceStrategy
(<code>string</code> alias)</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.Tracing">Tracing</a>)
</p>
<div>
<p>TraceStrategy defines the tracing strategy.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;parent&#34;</p></td>
<td><p>TraceStrategyParent enables tracing and only records spans if the parent span was sampled.</p>
</td>
</tr><tr><td><p>&#34;ratio&#34;</p></td>
<td><p>TraceStrategyRatio enables ratio-based tracing, defaulting to 100% sampling rate.</p>
</td>
</tr></tbody>
</table>
<h3 id="gateway.nginx.org/v1alpha1.Tracing">Tracing
</h3>
<p>
(<em>Appears on: </em><a href="#gateway.nginx.org/v1alpha1.ObservabilityPolicySpec">ObservabilityPolicySpec</a>)
</p>
<div>
<p>Tracing allows for enabling and configuring OpenTelemetry tracing.</p>
</div>
<table class="table table-bordered table-striped">
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>strategy</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.TraceStrategy">
TraceStrategy
</a>
</em>
</td>
<td>
<p>Strategy defines if tracing is ratio-based or parent-based.</p>
</td>
</tr>
<tr>
<td>
<code>ratio</code><br/>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ratio is the percentage of traffic that should be sampled. Integer from 0 to 100.
By default, 100% of http requests are traced. Not applicable for parent-based tracing.
If ratio is set to 0, tracing is disabled.</p>
</td>
</tr>
<tr>
<td>
<code>context</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.TraceContext">
TraceContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Context specifies how to propagate traceparent/tracestate headers.
Default: <a href="https://nginx.org/en/docs/ngx_otel_module.html#otel_trace_context">https://nginx.org/en/docs/ngx_otel_module.html#otel_trace_context</a></p>
</td>
</tr>
<tr>
<td>
<code>spanName</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>SpanName defines the name of the Otel span. By default is the name of the location for a request.
If specified, applies to all locations that are created for a route.
Format: must have all &lsquo;&ldquo;&rsquo; escaped and must not contain any &lsquo;$&rsquo; or end with an unescaped &lsquo;\&rsquo;
Examples of invalid names: some-$value, quoted-&ldquo;value&rdquo;-name, unescaped</p>
</td>
</tr>
<tr>
<td>
<code>spanAttributes</code><br/>
<em>
<a href="#gateway.nginx.org/v1alpha1.SpanAttribute">
[]SpanAttribute
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SpanAttributes are custom key/value attributes that are added to each span.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
</em></p>
