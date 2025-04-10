package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelloWorldShouldWork(t *testing.T) {
	message := "Hello World"
	require.EqualValues(t, "Hello World", message)
}
