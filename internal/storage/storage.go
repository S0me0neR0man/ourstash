// Package storage sync.map based storage
package storage

import (
	"errors"
	"sync"
	"sync/atomic"
)

const (
	metadataSection SectionIdType = 0
	dataSection     SectionIdType = 1

	// reserved for system unit
	sysUnitName string     = "sys"
	sysUnitId   UnitIdType = 0
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

// Storager interface key-value storage
type Storager interface {
}

// SyncMapStorage the in-memory NoSQL key-value tread safe storage.
// Most important part - synthetic key (see Key type)
type SyncMapStorage struct {
	m sync.Map

	// in this version only one data section (section = 1)
	lastRecId atomic.Uint64
}

// NewStorage create new storage
func NewStorage() (*SyncMapStorage, error) {
	s := SyncMapStorage{}
	s.lastRecId.Store(1)

	return &s, nil
}

func (s *SyncMapStorage) New(rec *Record) {
	recordId := s.lastRecId.Add(1)
	rec.put(recordId)
}
