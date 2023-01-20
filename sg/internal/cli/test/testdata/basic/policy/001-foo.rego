package main

deny_foo[msg] {
	input.name = "foo"

	msg = "name cannot be foo"
}

warn_foo[msg] {
	input.name = "foo"

	msg = "name is foo"
}

exception[rules] {
	input.skipped = true
	rules = ["foo"]
}
