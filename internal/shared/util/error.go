package util

import (
	"context"
	"sync"
)

type ErrorHandler struct {
	mu          sync.Mutex
	once        sync.Once
	combinedErr error
	cancel      context.CancelFunc
}

func NewErrorHandler(ctx context.Context) (context.Context, *ErrorHandler) {
	ctx, cancel := context.WithCancel(ctx)
	return ctx, &ErrorHandler{
		cancel: cancel,
	}
}

func (h *ErrorHandler) SetError(err error) {
	if err == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.combinedErr == nil {
		h.combinedErr = err
		h.once.Do(h.cancel)
	}
}

func (h *ErrorHandler) Err() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.combinedErr
}
