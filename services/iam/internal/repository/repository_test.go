package repository

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttributesJSONHelpers(t *testing.T) {
	t.Parallel()

	normalized := normalizeAttributes(nil)
	assert.Empty(t, normalized)

	raw, err := toJSON(map[string]any{"level": 3, "team": "search"})
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, "search", decoded["team"])
	assert.Equal(t, float64(3), decoded["level"])

	attrs, err := fromJSON(raw)
	require.NoError(t, err)
	assert.Equal(t, "search", attrs["team"])
	assert.Equal(t, float64(3), attrs["level"])

	attrs, err = fromJSON(nil)
	require.NoError(t, err)
	assert.Empty(t, attrs)
}

func TestNewRepo(t *testing.T) {
	t.Parallel()

	repo := NewRepo(nil)
	require.NotNil(t, repo)
	assert.Nil(t, repo.pool)
}

func TestAttributesJSONHelpersRejectUnsupportedValue(t *testing.T) {
	t.Parallel()

	_, err := toJSON(map[string]any{"bad": func() {}})
	require.Error(t, err)

	_, err = fromJSON([]byte("{"))
	require.Error(t, err)
}
