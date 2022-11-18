package main

import data.library

violation_no_baz[msg] {
	library.is_baz

	msg := "baz is not allowed"
}
