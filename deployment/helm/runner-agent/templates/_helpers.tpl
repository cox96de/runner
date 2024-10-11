{{- define "executor.namespace" -}}
{{- if eq "vm" .Values.engine}}{{.Values.vm.executorNamespace}}{{- end }}
{{- if eq "kube" .Values.engine}}{{.Values.kube.executorNamespace}}{{- end }}
{{- end}}
