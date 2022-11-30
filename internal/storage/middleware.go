package storage

// MiddlewarePutFunc is a function which receives an PutHandler and returns another PutHandler
type MiddlewarePutFunc func(handler PutHandler) PutHandler

// putMiddleware interface is anything which implements a MiddlewarePutFunc named Middleware
type putMiddleware interface {
	Middleware(handler PutHandler) PutHandler
}

// Middleware allows MiddlewarePutFunc to implement the putMiddleware interface
func (mw MiddlewarePutFunc) Middleware(handler PutHandler) PutHandler {
	return mw(handler)
}
