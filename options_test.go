package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_options_toMap_empty(t *testing.T) {
	t.Parallel()

	opts := new(options)

	data, err := opts.toMap()

	assert.NoError(t, err)
	assert.Empty(t, data)
}

func Test_options_toMap_withValues(t *testing.T) {
	t.Parallel()

	opts := &options{
		Kind:         javascript,
		Name:         "hitchhiker",
		PrimaryClass: "Guide",
		RepoName:     "xk6-hitchhiker",
	}

	data, err := opts.toMap()

	expected := map[string]interface{}{
		"kind":         string(javascript),
		"name":         "hitchhiker",
		"PrimaryClass": "Guide",
		"repoName":     "xk6-hitchhiker",
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, data)
}
