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

// putMiddlewarer interface is anything which implements a MiddlewarePutFunc named Middleware
type putMiddlewarer interface {
	PutMiddleware(PutHandler) PutHandler
}

// PutMiddleware allows MiddlewarePutFunc to implement the PutMiddleware interface
func (mw MiddlewarePutFunc) PutMiddleware(h PutHandler) PutHandler {
	return mw(h)
}

// PutChain use pattern chain of responsibility to save m
type PutChain struct {
	putMiddlewares []putMiddlewarer
}

// NewPutChain make new chain
func NewPutChain() (*PutChain, error) {
	sys, err := NewSysUnit()
	if err != nil {
		return nil, err
	}

	auth, err := NewAuthUnit()
	if err != nil {
		return nil, err
	}

	p := PutChain{}
	p.Attach(sys.PutMiddleware, auth.PutMiddleware)

	return &p, nil
}

// Attach appends a MiddlewarePutFunc to the put chain
func (p *PutChain) Attach(mwf ...MiddlewarePutFunc) *PutChain {
	for _, fn := range mwf {
		p.putMiddlewares = append(p.putMiddlewares, fn)
	}
	return p
}

func (p *PutChain) put(store Storager) error {
	if len(p.putMiddlewares) == 0 {
		return nil
	}

	var h PutHandler
	// Build PutMiddleware chain if no error was found
	for i := len(p.putMiddlewares) - 1; i >= 0; i-- {
		h = p.putMiddlewares[i].PutMiddleware(h)
	}

	return h.Put(store)
}
