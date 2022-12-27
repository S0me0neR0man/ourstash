// Package storage sync.map based storage
package storage

import (
	"sync"
)

var ()

// Storager interface key-value storage
type Storager interface {
	User() string
	Put(FieldIdType, any) error
}

// MyStorage the in-memory NoSQL key-value tread safe storage.
// Most important part - synthetic key (see Key type)
type MyStorage struct {
	m sync.Map
}

// NewStorage create new storage
func NewStorage() (*MyStorage, error) {
	return &MyStorage{}, nil
}
