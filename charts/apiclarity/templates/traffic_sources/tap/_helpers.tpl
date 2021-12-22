{{/* vim: set filetype=mustache: */}}
{{/*
Name of the chart.
*/}}
{{- define "apiclarity-taper.name" -}}
{{- printf "%s-%s-%s" .Release.Name .Chart.Name "taper" -}}
{{- end -}}

{{/*
Helm labels.
*/}}
{{- define "apiclarity-taper.labels" -}}
    app.kubernetes.io/name: {{ include "apiclarity.name" . }}-taper
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
{{- end -}}
