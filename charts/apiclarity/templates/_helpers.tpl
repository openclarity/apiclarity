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

{{/*
Create the name of the service account to use
*/}}
{{- define "apiclarity.serviceAccountName" -}}
{{- if .Values.apiclarity.serviceAccount.create -}}
    {{ default (include "apiclarity.name" .) .Values.apiclarity.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.apiclarity.serviceAccount.name }}
{{- end -}}
{{- end -}}
