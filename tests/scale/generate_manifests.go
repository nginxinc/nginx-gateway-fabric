package scale

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

var gwTmplTxt = `apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: gateway
spec:
  gatewayClassName: nginx
  listeners:
{{- range $l := . }}
  - name: {{ $l.Name }}
    hostname: "{{ $l.HostnamePrefix }}.example.com"{{ if ne $l.SecretName "" }}
    port: 443
    protocol: HTTPS
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        name: {{ $l.SecretName }}{{ else }}
    port: 80
    protocol: HTTP
	{{- end -}}
{{- end -}}`

var hrTmplTxt = `apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: {{ .Name }}
spec:
  parentRefs:
  - name: gateway
    sectionName: {{ .ListenerName }}
  hostnames:
  - "{{ .HostnamePrefix }}.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: {{ .BackendName }}
      port: 80`

// nolint:all
var secretTmplTxt = `apiVersion: v1
kind: Secret
metadata:
  name: {{ . }}
type: kubernetes.io/tls
data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNzakNDQVpvQ0NRQzdCdVdXdWRtRkNEQU5CZ2txaGtpRzl3MEJBUXNGQURBYk1Sa3dGd1lEVlFRRERCQmoKWVdabExtVjRZVzF3YkdVdVkyOXRNQjRYRFRJeU1EY3hOREl4TlRJek9Wb1hEVEl6TURjeE5ESXhOVEl6T1ZvdwpHekVaTUJjR0ExVUVBd3dRWTJGbVpTNWxlR0Z0Y0d4bExtTnZiVENDQVNJd0RRWUpLb1pJaHZjTkFRRUJCUUFECmdnRVBBRENDQVFvQ2dnRUJBTHFZMnRHNFc5aStFYzJhdnV4Q2prb2tnUUx1ek10U1Rnc1RNaEhuK3ZRUmxIam8KVzFLRnMvQVdlS25UUStyTWVKVWNseis4M3QwRGtyRThwUisxR2NKSE50WlNMb0NEYUlRN0Nhck5nY1daS0o4Qgo1WDNnVS9YeVJHZjI2c1REd2xzU3NkSEQ1U2U3K2Vab3NPcTdHTVF3K25HR2NVZ0VtL1Q1UEMvY05PWE0zZWxGClRPL051MStoMzROVG9BbDNQdTF2QlpMcDNQVERtQ0thaEROV0NWbUJQUWpNNFI4VERsbFhhMHQ5Z1o1MTRSRzUKWHlZWTNtdzZpUzIrR1dYVXllMjFuWVV4UEhZbDV4RHY0c0FXaGRXbElweHlZQlNCRURjczN6QlI2bFF1OWkxZAp0R1k4dGJ3blVmcUVUR3NZdWxzc05qcU95V1VEcFdJelhibHhJZVVDQXdFQUFUQU5CZ2txaGtpRzl3MEJBUXNGCkFBT0NBUUVBcjkrZWJ0U1dzSnhLTGtLZlRkek1ISFhOd2Y5ZXFVbHNtTXZmMGdBdWVKTUpUR215dG1iWjlpbXQKL2RnWlpYVE9hTElHUG9oZ3BpS0l5eVVRZVdGQ2F0NHRxWkNPVWRhbUloOGk0Q1h6QVJYVHNvcUNOenNNLzZMRQphM25XbFZyS2lmZHYrWkxyRi8vblc0VVNvOEoxaCtQeDljY0tpRDZZU0RVUERDRGh1RUtFWXcvbHpoUDJVOXNmCnl6cEJKVGQ4enFyM3paTjNGWWlITmgzYlRhQS82di9jU2lyamNTK1EwQXg4RWpzQzYxRjRVMTc4QzdWNWRCKzQKcmtPTy9QNlA0UFlWNTRZZHMvRjE2WkZJTHFBNENCYnExRExuYWRxamxyN3NPbzl2ZzNnWFNMYXBVVkdtZ2todAp6VlZPWG1mU0Z4OS90MDBHUi95bUdPbERJbWlXMGc9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
  tls.key: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2UUlCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktjd2dnU2pBZ0VBQW9JQkFRQzZtTnJSdUZ2WXZoSE4KbXI3c1FvNUtKSUVDN3N6TFVrNExFeklSNS9yMEVaUjQ2RnRTaGJQd0ZuaXAwMFBxekhpVkhKYy92TjdkQTVLeApQS1VmdFJuQ1J6YldVaTZBZzJpRU93bXF6WUhGbVNpZkFlVjk0RlAxOGtSbjl1ckV3OEpiRXJIUncrVW51L25tCmFMRHF1eGpFTVBweGhuRklCSnYwK1R3djNEVGx6TjNwUlV6dnpidGZvZCtEVTZBSmR6N3Rid1dTNmR6MHc1Z2kKbW9RelZnbFpnVDBJek9FZkV3NVpWMnRMZllHZWRlRVJ1VjhtR041c09va3R2aGxsMU1udHRaMkZNVHgySmVjUQo3K0xBRm9YVnBTS2NjbUFVZ1JBM0xOOHdVZXBVTHZZdFhiUm1QTFc4SjFINmhFeHJHTHBiTERZNmpzbGxBNlZpCk0xMjVjU0hsQWdNQkFBRUNnZ0VBQnpaRE50bmVTdWxGdk9HZlFYaHRFWGFKdWZoSzJBenRVVVpEcUNlRUxvekQKWlV6dHdxbkNRNlJLczUyandWNTN4cU9kUU94bTNMbjNvSHdNa2NZcEliWW82MjJ2dUczYnkwaVEzaFlsVHVMVgpqQmZCcS9UUXFlL2NMdngvSkczQWhFNmJxdFRjZFlXeGFmTmY2eUtpR1dzZk11WVVXTWs4MGVJVUxuRmZaZ1pOCklYNTlSOHlqdE9CVm9Sa3hjYTVoMW1ZTDFsSlJNM3ZqVHNHTHFybmpOTjNBdWZ3ZGRpK1VDbGZVL2l0K1EvZkUKV216aFFoTlRpNVFkRWJLVStOTnYvNnYvb2JvandNb25HVVBCdEFTUE05cmxFemIralQ1WHdWQjgvLzRGY3VoSwoyVzNpcjhtNHVlQ1JHSVlrbGxlLzhuQmZ0eVhiVkNocVRyZFBlaGlPM1FLQmdRRGlrR3JTOTc3cjg3Y1JPOCtQClpoeXltNXo4NVIzTHVVbFNTazJiOTI1QlhvakpZL2RRZDVTdFVsSWE4OUZKZnNWc1JRcEhHaTFCYzBMaTY1YjIKazR0cE5xcVFoUmZ1UVh0UG9GYXRuQzlPRnJVTXJXbDVJN0ZFejZnNkNQMVBXMEg5d2hPemFKZUdpZVpNYjlYTQoybDdSSFZOcC9jTDlYbmhNMnN0Q1lua2Iwd0tCZ1FEUzF4K0crakEyUVNtRVFWNXA1RnRONGcyamsyZEFjMEhNClRIQ2tTazFDRjhkR0Z2UWtsWm5ZbUt0dXFYeXNtekJGcnZKdmt2eUhqbUNYYTducXlpajBEdDZtODViN3BGcVAKQWxtajdtbXI3Z1pUeG1ZMXBhRWFLMXY4SDNINGtRNVl3MWdrTWRybVJHcVAvaTBGaDVpaGtSZS9DOUtGTFVkSQpDcnJjTzhkUVp3S0JnSHA1MzRXVWNCMVZibzFlYStIMUxXWlFRUmxsTWlwRFM2TzBqeWZWSmtFb1BZSEJESnp2ClIrdzZLREJ4eFoyWmJsZ05LblV0YlhHSVFZd3lGelhNcFB5SGxNVHpiZkJhYmJLcDFyR2JVT2RCMXpXM09PRkgKcmppb21TUm1YNmxhaDk0SjRHU0lFZ0drNGw1SHhxZ3JGRDZ2UDd4NGRjUktJWFpLZ0w2dVJSSUpBb0dCQU1CVApaL2p5WStRNTBLdEtEZHUrYU9ORW4zaGxUN3hrNXRKN3NBek5rbWdGMU10RXlQUk9Xd1pQVGFJbWpRbk9qbHdpCldCZ2JGcXg0M2ZlQ1Z4ZXJ6V3ZEM0txaWJVbWpCTkNMTGtYeGh3ZEVteFQwVit2NzZGYzgwaTNNYVdSNnZZR08KditwVVovL0F6UXdJcWZ6dlVmV2ZxdStrMHlhVXhQOGNlcFBIRyt0bEFvR0FmQUtVVWhqeFU0Ym5vVzVwVUhKegpwWWZXZXZ5TW54NWZyT2VsSmRmNzlvNGMvMHhVSjh1eFBFWDFkRmNrZW96dHNpaVFTNkN6MENRY09XVWxtSkRwCnVrdERvVzM3VmNSQU1BVjY3NlgxQVZlM0UwNm5aL2g2Tkd4Z28rT042Q3pwL0lkMkJPUm9IMFAxa2RjY1NLT3kKMUtFZlNnb1B0c1N1eEpBZXdUZmxDMXc9Ci0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K
`

var appTmplTxt = `apiVersion: v1
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ . }}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{ . }}
  template:
    metadata:
      labels:
        app: {{ . }}
    spec:
      containers:
      - name: nginx
        image: nginxdemos/nginx-hello:plain-text
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: {{ . }}
spec:
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: {{ . }}
`

var (
	gwTmpl     = template.Must(template.New("gw").Parse(gwTmplTxt))
	hrTmpl     = template.Must(template.New("hr").Parse(hrTmplTxt))
	secretTmpl = template.Must(template.New("secret").Parse(secretTmplTxt))
	appTmpl    = template.Must(template.New("app").Parse(appTmplTxt))
)

type Listener struct {
	Name           string
	HostnamePrefix string
	SecretName     string
}

type Route struct {
	Name           string
	ListenerName   string
	HostnamePrefix string
	BackendName    string
}

func getPrereqDirName(manifestDir string) string {
	return filepath.Join(manifestDir, "prereqs")
}

func generateScaleListenerManifests(numListeners int, manifestDir string, tls bool) error {
	listeners := make([]Listener, 0)
	backends := make([]string, 0)
	secrets := make([]string, 0)

	for i := 0; i < numListeners; i++ {
		listenerName := fmt.Sprintf("listener-%d", i)
		hostnamePrefix := fmt.Sprintf("%d", i)
		backendName := fmt.Sprintf("backend-%d", i)

		var secretName string
		if tls {
			secretName = fmt.Sprintf("secret-%d", i)
			secrets = append(secrets, secretName)
		}

		listeners = append(listeners, Listener{
			Name:           listenerName,
			HostnamePrefix: hostnamePrefix,
			SecretName:     secretName,
		})

		route := Route{
			Name:           fmt.Sprintf("route-%d", i),
			ListenerName:   listenerName,
			HostnamePrefix: hostnamePrefix,
			BackendName:    backendName,
		}

		backends = append(backends, backendName)

		if err := generateManifests(manifestDir, i, listeners, []Route{route}); err != nil {
			return err
		}
	}

	if err := generateSecrets(getPrereqDirName(manifestDir), secrets); err != nil {
		return err
	}

	return generateBackendAppManifests(getPrereqDirName(manifestDir), backends)
}

func generateSecrets(secretsDir string, secrets []string) error {
	err := os.Mkdir(secretsDir, 0o750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	for _, secret := range secrets {
		var buf bytes.Buffer

		if err = secretTmpl.Execute(&buf, secret); err != nil {
			return err
		}

		path := filepath.Join(secretsDir, fmt.Sprintf("%s.yaml", secret))

		fmt.Println("Writing", path)
		if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
			return err
		}
	}

	return nil
}

func generateScaleHTTPRouteManifests(numRoutes int, manifestDir string) error {
	l := Listener{
		Name:           "listener",
		HostnamePrefix: "*",
	}

	backendName := "backend"

	for i := 0; i < numRoutes; i++ {

		route := Route{
			Name:           fmt.Sprintf("route-%d", i),
			HostnamePrefix: fmt.Sprintf("%d", i),
			ListenerName:   "listener",
			BackendName:    backendName,
		}

		var listeners []Listener
		if i == 0 {
			// only generate a Gateway on the first iteration
			listeners = []Listener{l}
		}

		if err := generateManifests(manifestDir, i, listeners, []Route{route}); err != nil {
			return err
		}

	}

	return generateBackendAppManifests(getPrereqDirName(manifestDir), []string{backendName})
}

func generateManifests(outDir string, version int, listeners []Listener, routes []Route) error {
	var buf bytes.Buffer

	if len(listeners) > 0 {
		if err := gwTmpl.Execute(&buf, listeners); err != nil {
			return err
		}
	}

	for _, r := range routes {
		if buf.Len() > 0 {
			buf.Write([]byte("\n---\n"))
		}

		if err := hrTmpl.Execute(&buf, r); err != nil {
			return err
		}
	}

	err := os.Mkdir(outDir, 0o750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	filename := fmt.Sprintf("manifest-%d.yaml", version)
	path := filepath.Join(outDir, filename)

	fmt.Println("Writing", path)
	return os.WriteFile(path, buf.Bytes(), 0o600)
}

func generateBackendAppManifests(outDir string, backends []string) error {
	err := os.Mkdir(outDir, 0o750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	for _, backend := range backends {
		var buf bytes.Buffer

		if err = appTmpl.Execute(&buf, backend); err != nil {
			return err
		}

		path := filepath.Join(outDir, fmt.Sprintf("%s.yaml", backend))

		fmt.Println("Writing", path)
		if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
			return err
		}
	}

	return nil
}
