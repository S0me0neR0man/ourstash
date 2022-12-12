// Package storage sync.map based storage
package storage

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
)

const (
	metadataSection SectionIdType = 0
	dataSection     SectionIdType = 1

	sysUnitName  string = "sys"
	authUnitName string = "auth"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

// Storager interface key-value storage
type Storager interface {
	User() string
	Put(FieldIdType, any) error
}

// SyncMapStorage the in-memory NoSQL key-value tread safe storage.
// Most important part - synthetic key (see Key type)
type SyncMapStorage struct {
	m sync.Map
}

// NewStorage create new storage
func NewStorage() (*SyncMapStorage, error) {
	return &SyncMapStorage{}, nil
}

func (s *SyncMapStorage) getRecId(section SectionIdType) RecordIdType {
	key := NewKey(section, 0, 0, 0)
	id, _ := s.m.LoadOrStore(key, RecordIdType(1))
	return RecordIdType(reflect.ValueOf(id).Uint())
}

func (s *SyncMapStorage) getOrRegisterUnit(unit Uniter) (UnitIdType, error) {
	//key = NewKey(0, s.getRecId(metadataSection), 0, 0)
	//s.m.Store(key, unitName) // todo: unit id generator, and storage

	return 0, nil
}

func (s *SyncMapStorage) getUnitId(unit Uniter) (UnitIdType, error) {
	key := Key{}
	keyType := reflect.TypeOf(key)
	unitName := unit.Name()
	nameType := reflect.TypeOf(unitName)

	found := false
	var err error
	s.m.Range(func(k, v any) bool {
		if !reflect.ValueOf(k).CanConvert(keyType) {
			err = fmt.Errorf("on call Range() wrong key '%+v'", k)
			return false
		}
		key, err := NewKeyFromBytes(reflect.ValueOf(k).Bytes())
		if err != nil {
			return false
		}
		if key.Section() != 0 || key.Unit() != 0 {
			return true
		}
		if !reflect.ValueOf(v).CanConvert(nameType) {
			err = fmt.Errorf("on call Range() wrong unit name value '%+v'", k)
			return false
		}
		if unitName == reflect.ValueOf(v).String() {
			found = true
			return false
		}
		return true
	})

	if err != nil {
		err = fmt.Errorf("getOrRegisterUnit: %w", err)
		return 0, err
	}

	if !found {
		return 0, errors.New("not found")
	}

	return key.Unit(), nil
}

func (s *SyncMapStorage) New(chain *PutChain) error {
	log.Print("SyncMapStorage New()")
	store := newStoragerImpl(dataSection, s.getRecId(dataSection), s)

	// todo: goroutines pool
	if err := chain.put(store); err != nil {
		return err
	}
	return nil
}
