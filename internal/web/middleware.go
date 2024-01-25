package web

type Middleware func(Handler) Handler

// WrapMiddleware wraps a handler with middleware functions.
// The middleware functions are executed in reverse order.
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
