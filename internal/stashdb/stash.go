package stashdb

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

type OperationType string
type GUIDType string

const (
	metadataSection  SectionIdType = 0
	metadataRecordId RecordIdType  = 0
	counterFieldId   FieldIdType   = 0
	headerFieldId    FieldIdType   = 0

	InsertOperation OperationType = "insert"
	UpdateOperation OperationType = "update"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrFieldNotFound  = errors.New("field not found")
	ErrNotImplemented = errors.New("not implemented")
)

type recordHeader struct {
	guid      GUIDType
	next      RecordIdType
	operation OperationType
	// user      string
	time    time.Time
	deleted bool
}

func newRecordHeader(op OperationType) recordHeader {
	return recordHeader{
		guid:      GUIDType(uuid.New().String()),
		time:      time.Now(),
		deleted:   false,
		operation: op,
	}
}

type Record struct {
	guid GUIDType
	data map[string]any
}

// stash the in-memory NoSQL key-value tread safe stashdb.
// Most important part - synthetic key (see Key type)
type stash struct {
	redBlackTree
	m  sync.Map
	mu sync.RWMutex

	fields    map[SectionIdType]map[string]FieldIdType
	fieldsSFG singleflight.Group
	fieldsMu  sync.Mutex

	records    map[SectionIdType]map[GUIDType]Key
	recordsSFG singleflight.Group
	recordsMu  sync.Mutex

	sugar *zap.SugaredLogger
}

func newStash(logger *zap.Logger) *stash {
	return &stash{
		redBlackTree: redBlackTree{},
		sugar:        logger.Sugar(),
		fields:       make(map[SectionIdType]map[string]FieldIdType, 0),
		records:      make(map[SectionIdType]map[GUIDType]Key, 0),
	}
}

func (s *stash) newId(section SectionIdType) RecordIdType {
	key := NewKey(section, metadataRecordId, counterFieldId)
	var firstId uint64 = 1
	aid, loaded := s.m.LoadOrStore(key, &firstId)
	if !loaded {
		s.put(key) // todo: move to init section
		return RecordIdType(firstId)
	}
	return RecordIdType(atomic.AddUint64(aid.(*uint64), 1))
}

//func (s *stash) findRecord(section SectionIdType, record RecordIdType, field FieldIdType) (*redBlackNode, bool) {
//	it := s.iterator()
//	for flag := it.next(); flag; flag = it.next() {
//		if it.node.key.Section() < section || it.node.key.Record() < record {
//			continue
//		}
//		if it.node.key.Section() == section && it.node.key.Record() == record && it.node.key.Field() == field {
//			return it.node, true
//		}
//		if (it.node.key.Section() == section && it.node.key.Record() > record) || it.node.key.Section() > section {
//			return nil, false
//		}
//	}
//	return nil, false
//}

//func (s *stash) getStringValue(key Key) (string, error) {
//	const msg = "getStringValue:"
//	value, ok := s.m.Load(key)
//	if !ok {
//		return "", fmt.Errorf("%s key %s not found", msg, key.String())
//	}
//
//	var str string
//	str, ok = value.(string)
//	if !ok {
//		return "", fmt.Errorf("%s stored value is not string (key %s)", msg, key.String())
//	}
//
//	return str, nil
//}

func (s *stash) getRecordHeader(key Key) (recordHeader, error) {
	const msg = "getRecordHeader:"
	header := recordHeader{}
	value, ok := s.m.Load(key)
	if !ok {
		return header, fmt.Errorf("%s key %s not found", msg, key.String())
	}

	header, ok = value.(recordHeader)
	if !ok {
		return header, fmt.Errorf("%s stored value is not string (key %s)", msg, key.String())
	}

	return header, nil
}

func (s *stash) fieldIdSFG(section SectionIdType, fieldName string) FieldIdType {
	res, err, shared := s.fieldsSFG.Do(
		string(section)+fieldName,
		func() (interface{}, error) {
			s.fieldsMu.Lock()
			defer s.fieldsMu.Unlock()

			if s.fields[section] == nil {
				s.fields[section] = make(map[string]FieldIdType)
			}

			fid, ok := s.fields[section][fieldName]
			if !ok {
				fid = FieldIdType(len(s.fields[section]) + 1)
				key := NewKey(section, metadataRecordId, fid)
				s.m.Store(key, fieldName)
				s.put(key)
				s.fields[section][fieldName] = fid
			}
			return fid, nil
		})

	if err != nil {
		s.sugar.Errorw("fieldIdSFG", "res", res, "err", err, "shared", shared)
	}

	return res.(FieldIdType)
}

func (s *stash) fieldNameSFG(section SectionIdType, fieldId FieldIdType) (string, error) {
	res, err, shared := s.fieldsSFG.Do(
		string(section)+"-"+strconv.Itoa(int(fieldId)),
		func() (interface{}, error) {
			s.fieldsMu.Lock()
			if s.fields[section] == nil {
				s.fields[section] = make(map[string]FieldIdType) // todo: make all on start and remove lock
				s.fieldsMu.Unlock()
				return "", ErrFieldNotFound
			}
			s.fieldsMu.Unlock()

			for fname, fid := range s.fields[section] {
				if fid == fieldId {
					return fname, nil
				}
			}
			return "", ErrFieldNotFound
		})

	if err != nil {
		if !errors.Is(ErrFieldNotFound, err) {
			s.sugar.Errorw("fieldNameSFG", "res", res, "err", err, "shared", shared)
		}
		return "", err
	}
	return res.(string), nil
}

func (s *stash) recordKeySFG(section SectionIdType, guid GUIDType) (Key, error) {
	res, err, shared := s.recordsSFG.Do(
		string(section)+string(guid),
		func() (interface{}, error) {
			s.recordsMu.Lock()
			if s.records[section] == nil {
				s.records[section] = make(map[GUIDType]Key) // todo: make all on start and remove lock
				s.recordsMu.Unlock()
				return Key{}, ErrRecordNotFound
			}
			s.recordsMu.Unlock()

			recKey, ok := s.records[section][guid]
			if !ok {
				return Key{}, ErrRecordNotFound
			}
			return recKey, nil
		})

	if err != nil {
		if !errors.Is(ErrRecordNotFound, err) {
			s.sugar.Errorw("recordKeySFG", "res", res, "err", err, "shared", shared)
		}
		return res.(Key), err
	}
	return res.(Key), nil
}

func (s *stash) recordAddSFG(section SectionIdType, guid GUIDType, recKey Key) {
	res, err, shared := s.recordsSFG.Do(
		string(section)+string(guid),
		func() (interface{}, error) {
			s.recordsMu.Lock()
			defer s.recordsMu.Unlock()

			if s.records[section] == nil {
				s.records[section] = make(map[GUIDType]Key)
			}

			s.records[section][guid] = recKey
			return recKey, nil
		})

	if err != nil {
		s.sugar.Errorw("recordAddSFG", "res", res, "err", err, "shared", shared)
	}
}

func (s *stash) recordRemoveSFG(section SectionIdType, guid GUIDType) (Key, error) {
	res, err, shared := s.recordsSFG.Do(
		"remove"+string(section)+string(guid),
		func() (interface{}, error) {
			s.recordsMu.Lock()
			defer s.recordsMu.Unlock()

			if s.records[section] == nil {
				s.records[section] = make(map[GUIDType]Key)
			}

			recKey, ok := s.records[section][guid]
			if !ok {
				return Key{}, ErrRecordNotFound
			}
			delete(s.records[section], guid)
			return recKey, nil
		})

	if err != nil {
		if !errors.Is(ErrRecordNotFound, err) {
			s.sugar.Errorw("recordRemoveSFG", "res", res, "err", err, "shared", shared)
		}
		return res.(Key), err
	}
	return res.(Key), nil
}

func (s *stash) putHeader(section SectionIdType, f func() recordHeader) (GUIDType, RecordIdType) {
	recId := s.newId(section)
	key := NewKey(section, recId, headerFieldId)
	header := f()
	s.m.Store(key, header)
	s.recordAddSFG(section, header.guid, key)
	s.put(key)

	s.sugar.Debugw("put header", "operation", header.operation, "guid", header.guid, "key", key)
	return header.guid, recId
}

func (s *stash) putData(section SectionIdType, recId RecordIdType, data map[string]any) {
	for name, value := range data {
		fid := s.fieldIdSFG(section, name)
		key := NewKey(section, recId, fid)
		s.m.Store(key, value)
		s.put(key)
		s.sugar.Debugw("put data", "name", name, "key", key)
	}
}

// Insert data
func (s *stash) Insert(section SectionIdType, data map[string]any) GUIDType {
	s.mu.Lock()
	defer s.mu.Unlock()

	guid, recId := s.putHeader(section, func() recordHeader {
		header := newRecordHeader(InsertOperation)
		return header
	})
	s.putData(section, recId, data)

	return guid
}

// Get data
func (s *stash) Get(section SectionIdType, guid GUIDType) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, err := s.recordKeySFG(section, guid)
	if err != nil {
		return nil, err
	}
	s.sugar.Debugw("get", "guid", guid, "key", key)

	node := s.get(key)
	if node == nil {
		return nil, ErrRecordNotFound
	}
	recId := node.key.Record()

	res := make(map[string]any)
	it := s.iteratorAt(node)
	for it.pos == onmyway {
		if it.node.key.Section() != section || it.node.key.Record() != recId {
			break
		}
		if it.node.key.Field() == headerFieldId {
			it.next()
			continue
		}
		name, err := s.fieldNameSFG(section, it.node.key.Field())
		if err != nil {
			s.sugar.Debugw("fieldNameSFG", "err", err)
			return nil, err
		}
		value, ok := s.m.Load(it.node.key)
		if !ok {
			s.sugar.Debugw("s.m.Load", "err", "!ok")
			return nil, errors.New("get: impossible, value stolen")
		}
		res[name] = value

		it.next()
	}

	return res, nil
}

// Remove data
func (s *stash) Remove(section SectionIdType, guid GUIDType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, err := s.recordRemoveSFG(section, guid)
	if err != nil {
		return err
	}
	s.sugar.Debugw("remove", "guid", guid, "key", key)

	var header recordHeader
	header, err = s.getRecordHeader(key)
	if err != nil {
		return err
	}
	header.deleted = true
	s.m.Store(key, header)

	return nil
}

// Update data
func (s *stash) Update(section SectionIdType, guid GUIDType, data map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prevKey, err := s.recordKeySFG(section, guid)
	if err != nil {
		return err
	}

	var prevHeader recordHeader
	prevHeader, err = s.getRecordHeader(prevKey)
	if err != nil {
		return err
	}
	prevHeader.deleted = true
	prevHeader.time = time.Now()

	guid, recId := s.putHeader(section, func() recordHeader {
		header := newRecordHeader(UpdateOperation)
		header.guid = guid
		return header
	})
	s.putData(section, recId, data)

	prevHeader.next = recId
	s.m.Store(prevKey, prevHeader)

	s.sugar.Debugw("update", "guid", guid, "prevKey", prevKey)
	return nil
}

// Find data
func (s *stash) Find(ctx context.Context, section SectionIdType, f func(*map[string]any) bool) ([]Record, error) {
	var founded []Record

	s.recordsMu.Lock()
	if s.records[section] == nil {
		s.records[section] = make(map[GUIDType]Key) // todo: make all on start and remove lock
	}
	recordsInSection := s.records[section]
	s.recordsMu.Unlock()

	// todo: run s.Get in goroutines with context
	for guid := range recordsInSection {
		data, err := s.Get(section, guid)
		if err != nil && !errors.Is(ErrRecordNotFound, err) {
			return nil, err
		}
		if err != nil {
			s.sugar.Warnw("get", "err", err)
		}
		if f == nil || f(&data) {
			founded = append(founded, Record{
				guid: guid,
				data: data,
			})
		}
	}
	return founded, nil
}
