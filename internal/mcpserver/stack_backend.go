package mcpserver

import (
	"github.com/kiwifs/kiwifs/internal/bootstrap"
)

// stackBackend reuses a live bootstrap stack without closing it on MCP shutdown.
type stackBackend struct {
	*LocalBackend
}

// NewStackBackend reuses a live bootstrap stack (e.g. from kiwifs serve) without owning its lifetime.
func NewStackBackend(stack *bootstrap.Stack) Backend {
	return &stackBackend{LocalBackend: newLocalBackendFromStack(stack)}
}

func newLocalBackendFromStack(stack *bootstrap.Stack) *LocalBackend {
	b := &LocalBackend{root: stack.Root, stack: stack}
	b.once.Do(func() {}) // skip bootstrap rebuild; stack is injected
	return b
}

func (b *stackBackend) Close() error { return nil }

var _ Backend = (*stackBackend)(nil)
