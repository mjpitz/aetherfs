// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package dataset

import (
	"encoding"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

const tagSetPrefix = "json:"

func splitDatasetTag(dataset string) (tag Tag, err error) {
	parts := strings.Split(dataset, ":")

	switch {
	case len(parts) > 2:
		err = fmt.Errorf("too many parts")
	case len(parts) == 2:
		tag.Dataset = parts[0]
		tag.Version = parts[1]
	case len(parts) == 1:
		tag.Dataset = parts[0]
		tag.Version = "latest"
	}

	return tag, err
}

// Tag represents the components that make up a Tag flag.
type Tag struct {
	Host    string
	Dataset string
	Version string
}

func (t *Tag) UnmarshalText(text []byte) (err error) {
	url := string(text)
	tmp := Tag{}

	parts := strings.Split(url, "/")

	switch {
	case len(parts) > 3:
		err = fmt.Errorf("too many parts to path")
	case len(parts) == 3:
		tmp, err = splitDatasetTag(strings.Join(parts[1:], "/"))
		if err == nil {
			tmp.Host = parts[0]
		}
	case len(parts) == 2 && !strings.HasPrefix(parts[0], "@"):
		tmp, err = splitDatasetTag(parts[1])
		if err == nil {
			tmp.Host = parts[0]
		}

	case len(parts) <= 2:
		tmp, err = splitDatasetTag(url)
		if err == nil {
			tmp.Host = "localhost:8080"
		}
	}

	if err != nil {
		return err
	}

	t.Host = tmp.Host
	t.Dataset = tmp.Dataset
	t.Version = tmp.Version

	return nil
}

func (t *Tag) MarshalText() (text []byte, err error) {
	return []byte(t.String()), nil
}

func (t *Tag) String() string {
	str := t.Dataset

	if len(t.Version) > 0 {
		str = str + ":" + t.Version
	}

	if len(t.Host) > 0 {
		str = t.Host + "/" + str
	}

	return str
}

var (
	_ encoding.TextMarshaler   = &Tag{}
	_ encoding.TextUnmarshaler = &Tag{}
	_ fmt.Stringer             = &Tag{}
)

// TagSet provides a custom data type that can
type TagSet struct {
	value []*Tag
}

func (t *TagSet) Set(value string) error {
	if strings.HasPrefix(value, tagSetPrefix) {
		value := strings.TrimPrefix(value, tagSetPrefix)

		return json.Unmarshal([]byte(value), &t.value)
	}

	tag := &Tag{}

	err := tag.UnmarshalText([]byte(value))
	if err != nil {
		return err
	}

	t.value = append(t.value, tag)
	return nil
}

func (t *TagSet) String() string {
	return fmt.Sprintf("%s", t.value)
}

func (t *TagSet) Serialize() string {
	data, _ := json.Marshal(t.value)

	return tagSetPrefix + string(data)
}

func (t *TagSet) Value() []*Tag {
	return append([]*Tag{}, t.value...)
}

func (t *TagSet) Len() int {
	return len(t.value)
}

var (
	_ cli.Generic    = &TagSet{}
	_ cli.Serializer = &TagSet{}
)

// ParseTag provides a convenience function for quickly converting a string tag into it's structural components.
func ParseTag(val string) (tag Tag, err error) {
	return tag, tag.UnmarshalText([]byte(val))
}
