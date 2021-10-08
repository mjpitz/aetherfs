// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

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
