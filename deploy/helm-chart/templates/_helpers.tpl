{{/*
Expand the name of the chart.
*/}}
{{- define "nginx-gateway.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "nginx-gateway.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "nginx-gateway.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "nginx-gateway.labels" -}}
helm.sh/chart: {{ include "nginx-gateway.chart" . }}
{{ include "nginx-gateway.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "nginx-gateway.selectorLabels" -}}
app.kubernetes.io/name: {{ include "nginx-gateway.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the ServiceAccount to use
*/}}
{{- define "nginx-gateway.serviceAccountName" -}}
{{- default (include "nginx-gateway.fullname" .) .Values.gateway.serviceAccount.name }}
{{- end }}

{{/*
Expand default NGINX conf ConfigMap name.
*/}}
{{- define "nginx-gateway.nginx-conf" -}}
{{- printf "%s-%s" (include "nginx-gateway.fullname" .) "conf" -}}
{{- end -}}

{{/*
Expand default njs-modules ConfigMap name.
*/}}
{{- define "nginx-gateway.njs-modules" -}}
{{- printf "%s-%s" (include "nginx-gateway.fullname" .) "njs-modules" -}}
{{- end -}}
