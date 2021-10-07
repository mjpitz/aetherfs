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

package blocks

type Size int32

const (
	Byte Size = 1 << (10 * iota)
	Kibibyte
	Mebibyte
)

var (
	// PartSize is a cache-optimized length that is used to send and share parts of a block amongst a group of nodes.
	// It is also used during uploads and downloads as the segment sizes to avoid buffering gigabytes of data in memory.
	PartSize = 64 * Kibibyte
)
