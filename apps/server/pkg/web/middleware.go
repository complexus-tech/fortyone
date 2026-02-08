package web

// Middleware is a function that wraps a Handler to add cross-cutting concerns
// such as logging, authentication, or metrics. Middleware functions are
// composable and execute in the order they are provided.
type Middleware func(Handler) Handler

// wrapMiddleware wraps a handler with middleware functions.
// The middleware functions are executed in reverse order so that the first
// middleware in the slice is the outermost wrapper (executes first on the way in,
// last on the way out).
func wrapMiddleware(mw []Middleware, handler Handler) Handler {
	if len(mw) == 0 {
		return handler
	}

	wrapped := handler
	for i := len(mw) - 1; i >= 0; i-- {
		mwFunc := mw[i]
		if mwFunc != nil {
			wrapped = mwFunc(wrapped)
		}
	}

	return wrapped
}
