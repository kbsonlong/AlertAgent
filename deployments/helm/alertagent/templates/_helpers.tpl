{{/*
Expand the name of the chart.
*/}}
{{- define "alertagent.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "alertagent.fullname" -}}
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
{{- define "alertagent.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "alertagent.labels" -}}
helm.sh/chart: {{ include "alertagent.chart" . }}
{{ include "alertagent.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "alertagent.selectorLabels" -}}
app.kubernetes.io/name: {{ include "alertagent.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "alertagent.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "alertagent.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
API labels
*/}}
{{- define "alertagent.api.labels" -}}
{{ include "alertagent.labels" . }}
app.kubernetes.io/component: api
{{- end }}

{{/*
API selector labels
*/}}
{{- define "alertagent.api.selectorLabels" -}}
{{ include "alertagent.selectorLabels" . }}
app.kubernetes.io/component: api
{{- end }}

{{/*
Worker labels
*/}}
{{- define "alertagent.worker.labels" -}}
{{ include "alertagent.labels" . }}
app.kubernetes.io/component: worker
{{- end }}

{{/*
Worker selector labels
*/}}
{{- define "alertagent.worker.selectorLabels" -}}
{{ include "alertagent.selectorLabels" . }}
app.kubernetes.io/component: worker
{{- end }}

{{/*
Rule Server labels
*/}}
{{- define "alertagent.ruleServer.labels" -}}
{{ include "alertagent.labels" . }}
app.kubernetes.io/component: rule-server
{{- end }}

{{/*
Rule Server selector labels
*/}}
{{- define "alertagent.ruleServer.selectorLabels" -}}
{{ include "alertagent.selectorLabels" . }}
app.kubernetes.io/component: rule-server
{{- end }}

{{/*
Database host
*/}}
{{- define "alertagent.database.host" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" (include "alertagent.fullname" .) }}
{{- else }}
{{- .Values.secrets.database.host }}
{{- end }}
{{- end }}

{{/*
Database port
*/}}
{{- define "alertagent.database.port" -}}
{{- if .Values.postgresql.enabled }}
{{- "5432" }}
{{- else }}
{{- .Values.secrets.database.port }}
{{- end }}
{{- end }}

{{/*
Redis host
*/}}
{{- define "alertagent.redis.host" -}}
{{- if .Values.redis.enabled }}
{{- printf "%s-redis-master" (include "alertagent.fullname" .) }}
{{- else }}
{{- .Values.secrets.redis.host }}
{{- end }}
{{- end }}

{{/*
Redis port
*/}}
{{- define "alertagent.redis.port" -}}
{{- if .Values.redis.enabled }}
{{- "6379" }}
{{- else }}
{{- .Values.secrets.redis.port }}
{{- end }}
{{- end }}

{{/*
MySQL host for rule server
*/}}
{{- define "alertagent.mysql.host" -}}
{{- if .Values.ruleServer.mysql.enabled }}
{{- printf "%s-rule-server-mysql" (include "alertagent.fullname" .) }}
{{- else }}
{{- "mysql" }}
{{- end }}
{{- end }}

{{/*
Image pull policy
*/}}
{{- define "alertagent.imagePullPolicy" -}}
{{- .Values.global.imagePullPolicy | default "IfNotPresent" }}
{{- end }}

{{/*
Image registry
*/}}
{{- define "alertagent.imageRegistry" -}}
{{- if .Values.global.imageRegistry }}
{{- printf "%s/" .Values.global.imageRegistry }}
{{- end }}
{{- end }}

{{/*
Namespace
*/}}
{{- define "alertagent.namespace" -}}
{{- .Values.global.namespace | default .Release.Namespace }}
{{- end }}

{{/*
Storage class
*/}}
{{- define "alertagent.storageClass" -}}
{{- if .Values.global.storageClass }}
{{- .Values.global.storageClass }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Common environment variables
*/}}
{{- define "alertagent.commonEnv" -}}
- name: APP_NAME
  value: {{ .Values.config.app.name | quote }}
- name: APP_VERSION
  value: {{ .Values.config.app.version | quote }}
- name: APP_ENVIRONMENT
  value: {{ .Values.config.app.environment | quote }}
- name: LOG_LEVEL
  value: {{ .Values.config.logging.level | quote }}
- name: LOG_FORMAT
  value: {{ .Values.config.logging.format | quote }}
- name: METRICS_ENABLED
  value: {{ .Values.config.metrics.enabled | quote }}
- name: METRICS_PORT
  value: {{ .Values.config.metrics.port | quote }}
- name: TRACING_ENABLED
  value: {{ .Values.config.tracing.enabled | quote }}
- name: HEALTH_ENABLED
  value: {{ .Values.config.health.enabled | quote }}
- name: DATABASE_HOST
  value: {{ include "alertagent.database.host" . | quote }}
- name: DATABASE_PORT
  value: {{ include "alertagent.database.port" . | quote }}
- name: DATABASE_NAME
  valueFrom:
    secretKeyRef:
      name: {{ include "alertagent.fullname" . }}-secrets
      key: database-name
- name: DATABASE_USERNAME
  valueFrom:
    secretKeyRef:
      name: {{ include "alertagent.fullname" . }}-secrets
      key: database-username
- name: DATABASE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "alertagent.fullname" . }}-secrets
      key: database-password
- name: REDIS_HOST
  value: {{ include "alertagent.redis.host" . | quote }}
- name: REDIS_PORT
  value: {{ include "alertagent.redis.port" . | quote }}
- name: REDIS_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "alertagent.fullname" . }}-secrets
      key: redis-password
- name: JWT_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ include "alertagent.fullname" . }}-secrets
      key: jwt-secret
{{- end }}

{{/*
Common volume mounts
*/}}
{{- define "alertagent.commonVolumeMounts" -}}
- name: config
  mountPath: /app/config
  readOnly: true
- name: tmp
  mountPath: /tmp
{{- end }}

{{/*
Common volumes
*/}}
{{- define "alertagent.commonVolumes" -}}
- name: config
  configMap:
    name: {{ include "alertagent.fullname" . }}-config
- name: tmp
  emptyDir: {}
{{- end }}

{{/*
Common security context
*/}}
{{- define "alertagent.securityContext" -}}
runAsNonRoot: true
runAsUser: 65534
runAsGroup: 65534
fsGroup: 65534
seccompProfile:
  type: RuntimeDefault
{{- end }}

{{/*
Container security context
*/}}
{{- define "alertagent.containerSecurityContext" -}}
allowPrivilegeEscalation: false
readOnlyRootFilesystem: true
capabilities:
  drop:
    - ALL
runAsNonRoot: true
runAsUser: 65534
runAsGroup: 65534
{{- end }}

{{/*
Common probe configuration
*/}}
{{- define "alertagent.livenessProbe" -}}
httpGet:
  path: /health
  port: http
initialDelaySeconds: 30
periodSeconds: 10
timeoutSeconds: 5
failureThreshold: 3
{{- end }}

{{- define "alertagent.readinessProbe" -}}
httpGet:
  path: /ready
  port: http
initialDelaySeconds: 5
periodSeconds: 5
timeoutSeconds: 3
failureThreshold: 3
{{- end }}

{{- define "alertagent.startupProbe" -}}
httpGet:
  path: /health
  port: http
initialDelaySeconds: 10
periodSeconds: 10
timeoutSeconds: 5
failureThreshold: 30
{{- end }}