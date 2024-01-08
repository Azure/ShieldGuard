# METADATA
# custom:
#  enforcementAction: deny
package requiredannotations

import future.keywords.if

violation[{"msg": msg}] {
	is_target

	required_annotations := ["acme/owning-team", "acme/owning-contact"]

	required_annotation := required_annotations[_]
	not input.review.object.metadata.annotations[required_annotation]
	msg := sprintf(
		"%s/%s missing required annotation: %s",
		[
			input.review.kind.kind,
			input.review.object.metadata.name,
			required_annotation,
		],
	)
}

is_target if {
	input.review.kind.kind = "Deployment"
}

is_target if {
	input.review.kind.kind = "StatefulSet"
}

is_target if {
	input.review.kind.kind = "DaemonSet"
}

is_target if {
	input.review.kind.kind = "ReplicaSet"
}

is_target if {
	input.review.kind.kind = "ReplicationController"
}

is_target if {
	input.review.kind.kind = "Job"
}

is_target if {
	input.review.kind.kind = "Pod"
}

is_target if {
	input.review.kind.kind = "Service"
}
