{{/*
Common labels
*/}}
{{- define "whosay.labels" -}}
app: {{ .Values.labels.app }}
environment: {{ .Values.labels.environment }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "whosay.selectorLabels" -}}
app: {{ .Values.labels.app }}
{{- end }}
