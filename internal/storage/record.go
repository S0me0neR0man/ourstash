package storage

import "log"

type Recorder interface {
}

type Record struct {
	putMiddlewares []putMiddleware
}

func NewRecord() *Record {
	return &Record{}
}

// UseInPut appends a MiddlewarePutFunc to the put chain
func (r *Record) UseInPut(mwf ...MiddlewarePutFunc) *Record {
	for _, fn := range mwf {
		r.putMiddlewares = append(r.putMiddlewares, fn)
	}
	return r
}

// useInterfaceInPut appends a putMiddleware to the put chain
func (r *Record) useInterfaceInPut(mw putMiddleware) {
	r.putMiddlewares = append(r.putMiddlewares, mw)
}

func (r *Record) put(id uint64) {
	if len(r.putMiddlewares) == 0 {
		return
	}
	sys, err := NewSysUnit()
	if err != nil {
		log.Fatalln(err)
	}
	handler := sys.PutHandler(nil)
	h := handler
	// Build putMiddleware chain if no error was found
	for i := len(r.putMiddlewares) - 1; i >= 0; i-- {
		h = r.putMiddlewares[i].Middleware(h)
	}
	handler.Put(new(Key))
}
