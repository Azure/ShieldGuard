package main

deny_app_name_required[msg] {
	input.apiVersion
	input.metadata
	not input.metadata.labels["app/name"]
	msg := sprintf("app name label app/name is required for %s/%s .", [input.kind, input.metadata.name])
}
