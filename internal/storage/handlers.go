package storage

type TypeEvent int

const (
	GetEvent TypeEvent = iota
	PutEvent
	DeleteEvent
)

// PutHandler put putMiddleware handle.
type PutHandler interface {
	Put(*Key)
}

// The PutHandlerFunc type is an adapter to allow the use of
// ordinary functions as handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type PutHandlerFunc func(key *Key)

// Put calls f(k).
func (f PutHandlerFunc) Put(k *Key) {
	f(k)
}
