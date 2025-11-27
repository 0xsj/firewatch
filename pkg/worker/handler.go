package worker

import "context"

// Handler processes jobs of a specific type.
type Handler interface {
	// Handle processes the job and returns an error if it fails.
	// The job will be retried if an error is returned and retries remain.
	Handle(ctx context.Context, job *Job) error
}

// HandlerFunc is an adapter to allow ordinary functions to be used as handlers.
type HandlerFunc func(ctx context.Context, job *Job) error

// Handle implements the Handler interface.
func (f HandlerFunc) Handle(ctx context.Context, job *Job) error {
	return f(ctx, job)
}

// Middleware wraps a handler to add cross-cutting concerns.
type Middleware func(Handler) Handler

// Chain applies middlewares to a handler in order.
// The first middleware is the outermost (executes first).
func Chain(h Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
