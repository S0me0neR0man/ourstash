package storage

// FindHandler Find FindMiddleware handle.
type FindHandler interface {
	Find(Storager) error
}

// The FindHandlerFunc type is an adapter to allow the use of
// ordinary functions as handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type FindHandlerFunc func(Storager) error

// Find calls f(k).
func (f FindHandlerFunc) Find(store Storager) error {
	return f(store)
}

// MiddlewareFindFunc is a function which receives an FindHandler and returns another FindHandler
type MiddlewareFindFunc func(FindHandler) FindHandler

// findMiddlewarer interface is anything which implements a MiddlewareFindFunc named Middleware
type findMiddlewarer interface {
	FindMiddleware(FindHandler) FindHandler
}

// FindMiddleware allows MiddlewareFindFunc to implement the FindMiddleware interface
func (mw MiddlewareFindFunc) FindMiddleware(h FindHandler) FindHandler {
	return mw(h)
}

// FindChain use pattern chain of responsibility to save m
type FindChain struct {
	findMiddlewares []findMiddlewarer
}

// NewFindChain make new chain
func NewFindChain() (*FindChain, error) {
	p := FindChain{}

	return &p, nil
}

// Attach appends a MiddlewareFindFunc to the Find chain
func (p *FindChain) Attach(mwf ...MiddlewareFindFunc) *FindChain {
	for _, fn := range mwf {
		p.findMiddlewares = append(p.findMiddlewares, fn)
	}
	return p
}

func (p *FindChain) Find(store Storager) error {
	if len(p.findMiddlewares) == 0 {
		return nil
	}

	var h FindHandler
	// Build FindMiddleware chain if no error was found
	for i := len(p.findMiddlewares) - 1; i >= 0; i-- {
		h = p.findMiddlewares[i].FindMiddleware(h)
	}

	return h.Find(store)
}
