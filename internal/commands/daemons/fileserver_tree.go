package daemons

import (
	"strings"
)

const (
	fileMask      = 0b1000
	tagMask       = 0b0100
	datasetMask   = 0b0010
	directoryMask = 0b0001
)

type FileServerNode struct {
	children map[string]*FileServerNode
	mode     int
}

func (n *FileServerNode) IsFile() bool {
	return n.mode&fileMask > 0
}

func (n *FileServerNode) IsTag() bool {
	return n.mode&tagMask > 0
}

func (n *FileServerNode) IsDataset() bool {
	return n.mode&datasetMask > 0
}

func (n *FileServerNode) IsDirectory() bool {
	return n.mode&directoryMask == 0
}

func (n *FileServerNode) insert(path string, mode int) {
	if n == nil {
		*n = FileServerNode{
			children: make(map[string]*FileServerNode),
			mode:     directoryMask,
		}
	}

	ptr := n
	parts := strings.Split(path, "/")
	for _, part := range parts {
		val := ptr.children[part]

		if val == nil {
			val = &FileServerNode{
				children: make(map[string]*FileServerNode),
				mode:     directoryMask,
			}
			ptr.children[part] = val
		}

		ptr = val
	}
	ptr.mode = mode
}
