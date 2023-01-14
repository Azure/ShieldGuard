package main

deny_name[msg] {
	input.name == "foo"
	msg := sprintf("%s is not allowed", [input.name])
}

deny_name[msg] {
	input.value == "bar"
	msg := sprintf("value %s is not allowed", [input.value])
}

deny_name[msg] {
	input.name == "bar"
	msg := sprintf("%s is not allowed", [input.name])
}

deny_name[msg] {
	input.name == "baz"
	msg := sprintf("%s is not allowed", [input.name])
}
