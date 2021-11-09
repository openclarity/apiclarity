{{/* vim: set filetype=mustache: */}}
{{/*
Name of the chart.
*/}}
{{- define "apiclarity.name" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name -}}
{{- end -}}

{{/*
Helm labels.
*/}}
{{- define "apiclarity.labels" -}}
    app.kubernetes.io/name: {{ include "apiclarity.name" . }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
