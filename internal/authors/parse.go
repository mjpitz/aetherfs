// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package authors

import (
	"strings"

	"github.com/urfave/cli/v2"
)

// Parse parses the contents of the authors file. The file is format based on Google's open source guidelines. For more
// information, see https://opensource.google/docs/releasing/authors/
func Parse(contents string) []*cli.Author {
	lines := strings.Split(contents, "\n")
	results := make([]*cli.Author, 0, len(lines))

	for _, line := range lines {
		idx := strings.Index(line, "#")
		if idx > -1 {
			line = line[:idx]
		}

		line := strings.TrimSpace(line)

		name := line
		email := ""

		start := strings.LastIndex(line, "<")
		end := strings.LastIndex(line, ">")

		// emails are optional
		if start > -1 && end > -1 {
			name = strings.TrimSpace(line[:start])
			email = line[start+1 : end]
		}

		if name != "" {
			results = append(results, &cli.Author{
				Name:  name,
				Email: email,
			})
		}
	}

	return results
}
