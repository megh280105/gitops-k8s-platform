{{- define "taskapi.name" -}}
{{- .Chart.Name }}
{{- end }}

{{- define "taskapi.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "taskapi.labels" -}}
app.kubernetes.io/name: {{ include "taskapi.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "taskapi.selectorLabels" -}}
app.kubernetes.io/name: {{ include "taskapi.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "taskapi.postgresHost" -}}
{{- printf "%s-postgresql" .Release.Name }}
{{- end }}

{{- define "taskapi.databaseURL" -}}
{{- printf "postgres://%s:%s@%s:5432/%s?sslmode=disable"
    .Values.postgresql.auth.username
    .Values.postgresql.auth.password
    (include "taskapi.postgresHost" .)
    .Values.postgresql.auth.database }}
{{- end }}
