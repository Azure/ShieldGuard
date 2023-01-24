package main

deny_name[msg] {
	input.name == "foo" #1 success
	msg := sprintf("%s is not allowed", [input.name])
}

warn_name[msg] {
	input.name == "foo" # 1 success
	msg := sprintf("%s is not allowed", [input.name])
}

deny_name[msg] { # 2 success
	input.name == "baz"
	msg := sprintf("%s is not allowed", [input.name])
}

deny_other[msg] { #1 success
	input.name == "foo"
	msg := sprintf("%s is not allowed", [input.name])
}
