---
title: "TLS Passthrough"
weight: 800
toc: true
docs: "DOCS-000"
---

Learn how to use TLSRoutes to configure TLS Passthrough load-balancing with NGINX Gateway Fabric.

## Overview

In this guide, we will show how to configure TLS passthrough for your application, using a [TLSRoute](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TLSRoute).

## Note on Gateway API Experimental Features

{{< important >}} TLSRoute is a Gateway API resource from the experimental release channel. {{< /important >}}

{{<include "installation/install-gateway-api-experimental-features.md" >}}

## Before you begin

- [Install]({{< relref "installation/" >}}) NGINX Gateway Fabric with experimental features enabled.
- Save the public IP address and port of NGINX Gateway Fabric into shell variables:

   ```text
   GW_IP=XXX.YYY.ZZZ.III
   GW_TLS_PORT=<port number>
   ```

{{< note >}}In a production environment, you should have a DNS record for the external IP address that is exposed, and it should refer to the hostname that the Gateway will forward for.{{< /note >}}

## Set up

Create the `secure-app` application by copying and pasting the following block into your terminal:

```yaml
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: secure-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: secure-app
  template:
    metadata:
      labels:
        app: secure-app
    spec:
      containers:
        - name: secure-app
          image: nginxdemos/nginx-hello:plain-text
          ports:
            - containerPort: 8443
          volumeMounts:
            - name: secret
              mountPath: /etc/nginx/ssl
              readOnly: true
            - name: config-volume
              mountPath: /etc/nginx/conf.d
      volumes:
        - name: secret
          secret:
            secretName: app-tls-secret
        - name: config-volume
          configMap:
            name: secure-config
---
apiVersion: v1
kind: Service
metadata:
  name: secure-app
spec:
  ports:
    - port: 8443
      targetPort: 8443
      protocol: TCP
      name: https
  selector:
    app: secure-app
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: secure-config
data:
  app.conf: |-
    server {
      listen 8443 ssl;
      listen [::]:8443 ssl;

      server_name app.example.com;

      ssl_certificate /etc/nginx/ssl/tls.crt;
      ssl_certificate_key /etc/nginx/ssl/tls.key;

      default_type text/plain;

      location / {
        return 200 "hello from pod \$hostname\n";
      }
    }
---
apiVersion: v1
kind: Secret
metadata:
  name: app-tls-secret
data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURGRENDQWZ3Q0NRQ3EzQWxhdnJiaWpqQU5CZ2txaGtpRzl3MEJBUXNGQURCTU1Rc3dDUVlEVlFRR0V3SlYKVXpFTE1Ba0dBMVVFQ0F3Q1EwRXhGakFVQmdOVkJBY01EVk5oYmlCR2NtRnVZMmx6WTI4eEdEQVdCZ05WQkFNTQpEMkZ3Y0M1bGVHRnRjR3hsTG1OdmJUQWVGdzB5TURBek1qTXlNekl3TkROYUZ3MHlNekF6TWpNeU16SXdORE5hCk1Fd3hDekFKQmdOVkJBWVRBbFZUTVFzd0NRWURWUVFJREFKRFFURVdNQlFHQTFVRUJ3d05VMkZ1SUVaeVlXNWoKYVhOamJ6RVlNQllHQTFVRUF3d1BZWEJ3TG1WNFlXMXdiR1V1WTI5dE1JSUJJakFOQmdrcWhraUc5dzBCQVFFRgpBQU9DQVE4QU1JSUJDZ0tDQVFFQTJCRXhZR1JPRkhoN2VPMVlxeCtWRHMzRzMrVEhyTEZULzdEUFFEQlkza3pDCi9oZlprWCt3OW1NNkQ1RU9uK2lpVlNhUWlQMm1aNFA3N29pR0dmd3JrNjJ0eEQ5cHphODM5NC9aSjF5Q0dXZ1QKK2NWUEVZbkxjQktzSTRMcktJZ21oWVIwUjNzWWRjR1JkSXJWUFZlNUVUQlk1Z1U0RGhhMDZOUEIraitmK0krWgphWGIvMlRBekJhNHozMWpIQzg2amVQeTFMdklGazFiY3I2cSsxRGR5eklxcWxkRDYvU3Q4Q2t3cDlOaDFCUGFhCktZZ1ZVd010UVBib2s1cFFmbVMrdDg4NHdSM0dTTEU4VkxRbzgyYnJhNUR3emhIamlzOTlJRGhzbUt0U3lWOXMKaWNJbXp5dHBnSXlhTS9zWEhRQU9KbVFJblFteWgyekd1WFhTQ0lkRGtRSURBUUFCTUEwR0NTcUdTSWIzRFFFQgpDd1VBQTRJQkFRQ0tsVkhOZ1k5VHZLaW9Xb0tvdllCdnNRMmYrcmFOOEJwdWNDcnRvRm15NUczcGIzU2lPTndaCkF2cnhtSm4vR3lsa3JKTHBpQVA1eUNBNGI2Y2lYMnRGa3pQRmhJVFZKRTVBeDlpaEF2WWZwTUFSdWVqM29HN2UKd0xwQk1iUnlGbHJYV29NWUVBMGxsV0JueHRQQXZYS2Y4SVZGYTRSSDhzV1JJSDB4M2hFdjVtQ3VUZjJTRTg0QwpiNnNjS3Z3MW9CQU5VWGxXRVZVYTFmei9rWWZBa1lrdHZyV2JUcTZTWGxodXRJYWY4WEYzSUMrL2x1b3gzZThMCjBBcEFQVE5sZ0JwOTkvcXMrOG9PMWthSmQ1TmV6TnlJeXhSdUtJMzlDWkxuQm9OYmkzdlFYY1NzRCtYU2lYT0cKcEVnTjNtci8xRms4OVZMSENhTnkyKzBqMjZ0eWpiclcKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRRFlFVEZnWkU0VWVIdDQKN1Zpckg1VU96Y2JmNU1lc3NWUC9zTTlBTUZqZVRNTCtGOW1SZjdEMll6b1BrUTZmNktKVkpwQ0kvYVpuZy92dQppSVlaL0N1VHJhM0VQMm5OcnpmM2o5a25YSUlaYUJQNXhVOFJpY3R3RXF3amd1c29pQ2FGaEhSSGV4aDF3WkYwCml0VTlWN2tSTUZqbUJUZ09GclRvMDhINlA1LzRqNWxwZHYvWk1ETUZyalBmV01jTHpxTjQvTFV1OGdXVFZ0eXYKcXI3VU4zTE1pcXFWMFByOUszd0tUQ24wMkhVRTlwb3BpQlZUQXkxQTl1aVRtbEIrWkw2M3p6akJIY1pJc1R4VQp0Q2p6WnV0cmtQRE9FZU9LejMwZ09HeVlxMUxKWDJ5SndpYlBLMm1Bakpveit4Y2RBQTRtWkFpZENiS0hiTWE1CmRkSUloME9SQWdNQkFBRUNnZ0VCQUxYaW16ODZrT1A0bkhBcTFPYVEyb2l3dndhQTczbTNlUytZSm84eFk4NFcKcmxyNXRzUWR5dGxPcEhTd05yQjBSQnNNTU1XeFNPQ0JJWlltUlVVZ200cGd2Uk9rRWl2OG9VOThQMkE4SnFTKwprWHBFRjVCNi84K2pXRmM0Z1Q4SWhlMEZtR0VJQllvelhYL08wejBsV0h4WXg2MHluWUoycU9vS1FKT3A5YjlsCmpiUVBkaC9mN2ErRWF0RzZNUFlrNG5xSEY3a0FzcmNsRXo2SGUvaEx6NmRkSTJ1N2RMRjB6QlN0QjM5WDFRZysKZ1JzTittOXg1S1FVTXYxMktvajdLc2hEelozOG5hSjd5bDgycGhBV1lGZzBOZHlzRlBRbmt0WmlNSUxOblFjNwpOeUt0cHNQaUxIRE9ha05hdEZLU2lOaUJrUk1lY1ZUMlJNMzMzUG54bFVFQ2dZRUEvYTY5MEEralU4VFJNbVZyCk4vRnlYWkxYa1c5b2NxVjBRbTA0TDMrSExybFNCTlRWSzk2U1pVT203VjViTzIxNmd4S2dJK3IwYm5kdE5GTUQKLzFncDhsdlJNcUlIeGZTeUo4SHpsSzViT0lnaUpxRGhzK3BKWTZmLytIVzZ1QkZyN3NGS3lxbVlIQlA0SC9BdApsT3lLeEVjMHFXazFlT2tCMWNNSGx0WDRwemtDZ1lFQTJncDhDVDVYWjNMSWRQN2M1SHpDS1YwczBYS1hGNmYyCkxzclhPVlZaTmJCN1NIS1NsOTBIU2VWVGx3czdqSnNxcC9yWFY2aHF0eUdEaTg4aTFZekthcEF6dXl3b0U3TnEKMUJpd2ZYSURQeTlPNUdGNXFYNXFUeENzSWNIcmo2Z21XMEZVQWhoS1lQcDRxd1JMdzFMZkJsd3U1VmhuN3I3ego0SkZBTEFpdlp4a0NnWUJicnpuKzVvZjdFSmtqQTdDYWlYTHlDczVLUzkrTi8rcGl6NktNMkNSOWFKRVNHZkhwClp3bTErNXRyRXIwYVgxajE0bGRxWTlKdjBrM3ZxVWs2a2h5bThUUk1mbThjeG5GVkdTMzF3SVpMaWpmOWlndkkKd0paQnBFaEkvaE83enVBWmJGYWhwR1hMVUJSUFJyalNxQ01IQ1UwcEpWTWtIZUtCNVhqcXRPNm5VUUtCZ0NJUAp6VHlzYm44TW9XQVZpSEJ4Uk91dFVKa1BxNmJZYUU3N0JSQkIwd1BlSkFRM1VjdERqaVh2RzFYWFBXQkR4VEFrCnNZdFNGZ214eEprTXJNWnJqaHVEbDNFLy9xckZOb1VYcmtxS2l4Tk4wcWMreXdDOWJPSVpHcXJUWG5jOHIzRkcKRFZlZWI5QWlrTU0ya3BkYTFOaHJnaS8xMVphb1lmVE0vQmRrNi9IUkFvR0JBSnFzTmFZYzE2clVzYzAzUEwybApXUGNzRnZxZGI3SEJyakVSRkhFdzQ0Vkt2MVlxK0ZWYnNNN1FTQVZ1V1llcGxGQUpDYzcrSEt1YjRsa1hRM1RkCndSajJLK2pOUzJtUXp1Y2hOQnlBZ1hXVnYveHhMZEE3NnpuWmJYdjl5cXhnTVVjTVZwZGRuSkxVZm9QVVZ1dTcKS0tlVVU3TTNIblRKUStrcldtbUxraUlSCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K
EOF
```

This will create the **secure-app** Service and a Deployment. The secure app is configured to serve HTTPS traffic on port 8443 for the host app.example.com. For TLS termination, a self-signed TLS certificate, with the common name `app.example.com`, and key are used. The app responds to a client's HTTPS requests with a simple text response "hello from pod $POD_HOSTNAME".

Run the following command to verify the resources were created:

```shell
kubectl get pods,svc
```

The output should include the **secure-app** pod and the **secure-app** Service:

```text
NAME                              READY   STATUS      RESTARTS   AGE
pod/secure-app-575785644-kzqf6    1/1     Running     0          12s

NAME                  TYPE        CLUSTER-IP        EXTERNAL-IP   PORT(S)    AGE
service/secure-app    ClusterIP   192.168.194.152   <none>        8443/TCP   12s
```

Create a Gateway. This will create a TLS listener with the hostname `*.example.com` and passthrough TLS mode. Copy and paste this into your terminal.

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: gateway
  namespace: default
spec:
  gatewayClassName: nginx
  listeners:
  - name: tls
    port: 443
    protocol: TLS
    hostname: "*.example.com"
    allowedRoutes:
      namespaces:
        from: All
      kinds:
        - kind: TLSRoute
    tls:
      mode: Passthrough
EOF
```

This Gateway will configure NGINX Gateway Fabric to accept TLS connections on port 443 and route them to the corresponding backend Services without decryption. The routing is done based on the SNI, which allows clients to specify a server name (like example.com) during the SSL handshake.

{{< note >}}It is possible to add an HTTPS listener on the same port that terminates TLS connections so long as the hostname does not overlap with the TLS listener hostname.{{< /note >}}

Create a TLSRoute that attaches to the Gateway and routes requests to `app.example.com` to the `secure-app` Service:

```yaml
kubectl apply -f - <<EOF
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TLSRoute
metadata:
  name: tls-secure-app-route
  namespace: default
spec:
  parentRefs:
  - name: gateway
    namespace: default
  hostnames:
  - "app.example.com"
  rules:
  - backendRefs:
    - name: secure-app
      port: 8443
EOF
```

{{< note >}}To route to a Service in a Namespace different from the TLSRoute Namespace, create a [ReferenceGrant](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1beta1.ReferenceGrant) to permit the cross-namespace reference. {{< /note >}}

## Send traffic

Using the external IP address and port for NGINX Gateway Fabric, send traffic to the `secure-app` application.

{{< note >}}If you have a DNS record allocated for `app.example.com`, you can send the request directly to that hostname, without needing to resolve.{{< /note >}}

Send a request to the `secure-app` Service on the TLS port with the `--insecure` flag. The `--insecure` flag is required because the `secure-app` is using self-signed certificates.

```shell
curl --resolve app.example.com:$GW_TLS_PORT:$GW_IP https://app.example.com:$GW_TLS_PORT --insecure -v
```

```text
Added app.example.com:8443:127.0.0.1 to DNS cache
* Hostname app.example.com was found in DNS cache
*   Trying 127.0.0.1:8443...
* Connected to app.example.com (127.0.0.1) port 8443
* ALPN: curl offers h2,http/1.1
* (304) (OUT), TLS handshake, Client hello (1):
* (304) (IN), TLS handshake, Server hello (2):
* (304) (IN), TLS handshake, Unknown (8):
* (304) (IN), TLS handshake, Certificate (11):
* (304) (IN), TLS handshake, CERT verify (15):
* (304) (IN), TLS handshake, Finished (20):
* (304) (OUT), TLS handshake, Finished (20):
* SSL connection using TLSv1.3 / AEAD-AES256-GCM-SHA384 / [blank] / UNDEF
* ALPN: server accepted http/1.1
* Server certificate:
*  subject: C=US; ST=CA; L=San Francisco; CN=app.example.com
*  start date: Mar 23 23:20:43 2020 GMT
*  expire date: Mar 23 23:20:43 2023 GMT
*  issuer: C=US; ST=CA; L=San Francisco; CN=app.example.com
*  SSL certificate verify result: self signed certificate (18), continuing anyway.
* using HTTP/1.x
> GET / HTTP/1.1
> Host: app.example.com:8443
> User-Agent: curl/8.6.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Server: nginx/1.27.0
< Date: Wed, 14 Aug 2024 20:41:21 GMT
< Content-Type: text/plain
< Content-Length: 43
< Connection: keep-alive
<
hello from pod secure-app-575785644-kzqf6
```

Note that the server certificate used to terminate the TLS connection has the subject common name of `app.example.com`. This is the server certificate that the `secure-app` is configured with and shows that the TLS connection was terminated by the `secure-app`, not NGINX Gateway Fabric.
