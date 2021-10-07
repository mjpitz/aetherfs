// AetherFS - A virtual file system for small to medium sized datasets (MB or GB, not TB or PB).
// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package authors

import (
	"strings"

	"github.com/urfave/cli/v2"
)

// Parse parses the contents of an AUTHORS file. An AUTHORS file is a plaintext file whose contents details the primary
// contributors to the project. Each line in the file contains either a comment (denoted by the pound symbol, "#") or
// an author. Each author line should contain the name of the contributor and an optional email address. The format for
// and author line is "name <email>". For more information, see the AUTHORS file for this project.
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
