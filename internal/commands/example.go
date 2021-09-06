// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package commands

import "strings"

// ExampleString formats a list of examples so that they display properly in the terminal. This function just pulls
// things out into a simple helper.
func ExampleString(examples ...string) string {
	return strings.Join(examples, "\n   ")
}
