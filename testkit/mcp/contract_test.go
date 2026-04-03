package mcp

import (
	"testing"

	"github.com/dpopsuev/jericho/work"
)

func TestServerContract_MockServer(t *testing.T) {
	RunServerContract(t, func() work.Server {
		return NewMockServer()
	})
}
