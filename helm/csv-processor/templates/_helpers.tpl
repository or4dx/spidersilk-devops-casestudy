{{/*
Expand the name of the release. Truncated to 63 chars (DNS label limit).
*/}}
{{- define "csv-processor.fullname" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels — applied to every resource managed by this chart.
helm.sh/chart includes the version so Helm can track upgrades.
app.kubernetes.io/* follow the standard K8s recommended label set.
*/}}
{{- define "csv-processor.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Values.image.tag | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels — used in Deployment.spec.selector and Service.spec.selector.
MUST be immutable after initial deploy; never include version or chart fields here.
*/}}
{{- define "csv-processor.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
