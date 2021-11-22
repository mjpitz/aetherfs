// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package dataset_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mjpitz/aetherfs/internal/dataset"
)

func TestTag(t *testing.T) {
	testCases := []struct {
		url     string
		host    string
		dataset string
		version string
		error   string
	}{
		{
			url:     "custom.domain/@scope/dataset:tag",
			host:    "custom.domain",
			dataset: "@scope/dataset",
			version: "tag",
		},
		{
			url:     "custom.domain/dataset:tag",
			host:    "custom.domain",
			dataset: "dataset",
			version: "tag",
		},
		{
			url:     "localhost:8080/@scope/dataset:tag",
			host:    "localhost:8080",
			dataset: "@scope/dataset",
			version: "tag",
		},
		{
			url:     "localhost:8080/dataset:tag",
			host:    "localhost:8080",
			dataset: "dataset",
			version: "tag",
		},
		{
			url:     "@scope/dataset:tag",
			host:    "localhost:8080",
			dataset: "@scope/dataset",
			version: "tag",
		},
		{
			url:     "dataset:tag",
			host:    "localhost:8080",
			dataset: "dataset",
			version: "tag",
		},
		{
			url:     "dataset",
			host:    "localhost:8080",
			dataset: "dataset",
			version: "latest",
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase.url)

		ref, err := dataset.ParseTag(testCase.url)
		if len(testCase.error) > 0 {
			require.Error(t, err)
			require.Equal(t, testCase.error, err.Error())
		} else {
			require.NoError(t, err)

			require.Equal(t, testCase.host, ref.Host)
			require.Equal(t, testCase.dataset, ref.Dataset)
			require.Equal(t, testCase.version, ref.Version)
		}
	}
}

func TestTagSet(t *testing.T) {
	tags := &dataset.TagSet{}

	require.NoError(t, tags.Set("custom.domain/@scope/dataset:tag"))
	require.NoError(t, tags.Set("custom.domain/dataset:tag"))
	require.NoError(t, tags.Set("localhost:8080/@scope/dataset:tag"))
	require.NoError(t, tags.Set("localhost:8080/dataset:tag"))

	expected := `json:["custom.domain/@scope/dataset:tag","custom.domain/dataset:tag","localhost:8080/@scope/dataset:tag","localhost:8080/dataset:tag"]`
	snapshot := tags.Serialize()
	require.Equal(t, expected, snapshot)

	require.NoError(t, tags.Set("@scope/dataset:tag"))
	require.NoError(t, tags.Set("dataset:tag"))
	require.NoError(t, tags.Set("dataset"))

	values := tags.Value()
	require.Len(t, values, 7)

	require.NoError(t, tags.Set(snapshot))

	values = tags.Value()
	require.Len(t, values, 4)
}
