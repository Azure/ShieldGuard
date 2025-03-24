package main

deny_must_be_true[msg] {
	resource := input.resources[_]
	resource.properties.mustBeTrue != true

	msg := "mustBeTrue is not set or set to false"
}
