// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package commands

import "strings"

// ExampleString formats a list of examples so that they display properly in the terminal. This function just pulls
// things out into a simple helper.
func ExampleString(examples ...string) string {
	return strings.Join(examples, "\n   ")
}
