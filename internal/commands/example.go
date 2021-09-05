package commands

import "strings"

func ExampleString(examples ...string) string {
	return strings.Join(examples, "\n   ")
}
