package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolvePids(t *testing.T) {
	pids, err := resolvePids("/usr/sbin/apache2 -k start")
	assert.Nil(t, err)

	fmt.Printf("Pids: %v\n", pids)
}
