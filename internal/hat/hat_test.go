package hat

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_dockerReplaceFrom(t *testing.T) {
	assert.Equal(t,
		[]byte("FROM ubuntu\nRUN echo hello"), DockerReplaceFrom([]byte("FROM debian\nRUN echo hello"),
			"ubuntu",
		),
	)
}
