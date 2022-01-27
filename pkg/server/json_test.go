package server_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// requireJsonEqual simplifies comparing two objects after marshaling them to JSON.
//
// In particular, this simplifies comparisons across some of the protocol types,
// which were primarily designed for one-way marshaling, and so use
// `interface{}` for some fields that can (presumably?) accept multiple types,
// which means the unmarshalled versions won't compare properly to hand-crafted
// test objects.
func requireJsonEqual(t testing.TB, expected, actual interface{}) {
	t.Helper()

	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err, "Could not marshal expected object")

	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err, "Could not marshal actual object")

	require.JSONEq(t, string(expectedJSON), string(actualJSON))
}
