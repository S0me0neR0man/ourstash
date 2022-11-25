// Package storage sync.map based storage
package storage

import (
	"sync"
)

type Storage struct {
	m sync.Map
}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) getOne() (any, error) {
	return nil, nil
}
