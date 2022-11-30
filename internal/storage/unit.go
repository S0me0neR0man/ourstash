package storage

type Uniter interface {
	Name() string
}

type ICanPut interface {
	PutHandler(next PutHandler) PutHandler
}
