package storage

// PutHandler put PutMiddleware handle.
type PutHandler interface {
	Put(Storager) error
}

// The PutHandlerFunc type is an adapter to allow the use of
// ordinary functions as handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type PutHandlerFunc func(Storager) error

// Put calls f(k).
func (f PutHandlerFunc) Put(store Storager) error {
	return f(store)
}

// MiddlewarePutFunc is a function which receives an PutHandler and returns another PutHandler
type MiddlewarePutFunc func(PutHandler) PutHandler

// PutMiddlewarer interface is anything which implements a MiddlewarePutFunc named Middleware
type PutMiddlewarer interface {
	PutMiddleware(PutHandler) PutHandler
}

// PutMiddleware allows MiddlewarePutFunc to implement the PutMiddleware interface
func (mw MiddlewarePutFunc) PutMiddleware(h PutHandler) PutHandler {
	return mw(h)
}

// PutChain use pattern chain of responsibility to save data
type PutChain struct {
	putMiddlewares []PutMiddlewarer
}

// NewPutChain make new chain
func NewPutChain() (*PutChain, error) {
	sys, err := NewSysUnit()
	if err != nil {
		return nil, err
	}

	mw := make([]PutMiddlewarer, 1)
	mw[0] = MiddlewarePutFunc(func(next PutHandler) PutHandler {
		return sys
	})
	return &PutChain{putMiddlewares: mw}, nil
}

// Use appends a MiddlewarePutFunc to the put chain
func (p *PutChain) Use(mwf ...MiddlewarePutFunc) *PutChain {
	for _, fn := range mwf {
		p.putMiddlewares = append(p.putMiddlewares, fn)
	}
	return p
}

func (p *PutChain) put(id uint64) error {
	if len(p.putMiddlewares) == 0 {
		return nil
	}

	var h PutHandler
	// Build PutMiddleware chain if no error was found
	for i := len(p.putMiddlewares) - 1; i >= 0; i-- {
		h = p.putMiddlewares[i].PutMiddleware(h)
	}

	return h.Put(nil)
}
