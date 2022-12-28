package storage

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

const (
	metadataSection SectionIdType = 0
	counterRecordId RecordIdType  = 0
	counterFieldId  FieldIdType   = 0
	headerFieldId   FieldIdType   = 0

	InsertOperation = "insert"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrFieldNotFound  = errors.New("field not found")
	ErrNotImplemented = errors.New("not implemented")
)

type recordHeader struct {
	guid      string
	next      RecordIdType
	operation string
	user      string
	time      time.Time
	deleted   bool
}

func newRecordHeader() recordHeader {
	return recordHeader{
		guid:      uuid.New().String(),
		time:      time.Now(),
		deleted:   false,
		operation: InsertOperation,
	}
}

// stash the in-memory NoSQL key-value tread safe storage.
// Most important part - synthetic key (see Key type)
type stash struct {
	redBlackTree
	m  sync.Map
	mu sync.RWMutex

	fields   map[SectionIdType]map[string]FieldIdType
	sfFields singleflight.Group

	records   map[SectionIdType]map[string]Key
	sfRecords singleflight.Group

	sugar *zap.SugaredLogger
}

func newStash(logger *zap.Logger) *stash {
	return &stash{
		redBlackTree: redBlackTree{},
		sugar:        logger.Sugar(),
		fields:       make(map[SectionIdType]map[string]FieldIdType, 0),
		records:      make(map[SectionIdType]map[string]Key, 0),
	}
}

func (s *stash) newId(section SectionIdType) RecordIdType {
	key := NewKey(section, counterRecordId, counterFieldId)
	var firstId uint64 = 1
	aid, loaded := s.m.LoadOrStore(key, &firstId)
	if !loaded {
		s.put(key) // todo: move to init section
		return RecordIdType(firstId)
	}
	return RecordIdType(atomic.AddUint64(aid.(*uint64), 1))
}

// findRecord find first node in record
//
// IMPORTANT: does not provide thread safety
func (s *stash) findRecord(section SectionIdType, record RecordIdType) (*redBlackNode, bool) {
	it := s.iterator()
	for flag := it.next(); flag; flag = it.next() {
		if it.node.key.Section() < section || it.node.key.Record() < record {
			continue
		}
		if it.node.key.Section() == section && it.node.key.Record() != counterRecordId && it.node.key.Record() == record {
			return it.node, true
		}
		if (it.node.key.Section() == section && it.node.key.Record() > record) || it.node.key.Section() > section {
			return nil, false
		}
	}
	return nil, false
}

// getStringValue get string value from sync.Map
func (s *stash) getStringValue(key Key) (string, error) {
	const msg = "getStringValue:"
	value, ok := s.m.Load(key)
	if !ok {
		return "", fmt.Errorf("%s key %s not found", msg, key.String())
	}

	var str string
	str, ok = value.(string)
	if !ok {
		return "", fmt.Errorf("%s stored value is not string (key %s)", msg, key.String())
	}

	return str, nil
}

// getStringValue get string value from sync.Map
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

func (s *stash) fieldIdSingleFlight(section SectionIdType, fieldName string) FieldIdType {
	key := string(section)
	res, _, _ := s.sfFields.Do(key,
		func() (interface{}, error) {
			if s.fields[section] == nil {
				// todo: move init and change single flight key
				s.fields[section] = s.fieldsInSection(section)
			}
			fid, ok := s.fields[section][fieldName]
			if !ok {
				fid = FieldIdType(len(s.fields[section]) + 1)
				key := NewKey(section, counterRecordId, fid)
				s.m.Store(key, fieldName)
				s.put(key)
				s.fields[section][fieldName] = fid
			}
			return fid, nil
		})

	return res.(FieldIdType)
}

func (s *stash) fieldsInSection(section SectionIdType) map[string]FieldIdType {
	res := make(map[string]FieldIdType)

	node, found := s.findRecord(section, counterRecordId)
	if !found {
		return res
	}

	it := s.iteratorAt(node)
	for it.pos == onmyway {
		if it.node.key.Section() != section || it.node.key.Record() != counterRecordId {
			break
		}
		fieldName, err := s.getStringValue(it.node.key)
		if err == nil {
			res[fieldName] = it.node.key.Field()
		}
		it.next()
	}

	return res
}

func (s *stash) fieldNameSingleFlight(section SectionIdType, fieldId FieldIdType) (string, error) {
	key := string(section)
	res, err, _ := s.sfFields.Do(key,
		func() (interface{}, error) {
			if s.fields[section] == nil {
				// todo: move init and change single flight key
				s.fields[section] = s.fieldsInSection(section)
			}
			for fname, fid := range s.fields[section] {
				if fid == fieldId {
					return fname, nil
				}
			}
			return 0, ErrFieldNotFound
		})

	if err != nil {
		return "", err
	}
	return res.(string), nil
}

func (s *stash) recordKeySingleFlight(section SectionIdType, guid string) (Key, error) {
	key := string(section)
	res, err, _ := s.sfRecords.Do(key,
		func() (interface{}, error) {
			if s.records[section] == nil {
				// todo: move init and change single flight key
				s.records[section] = s.recordsInSection(section)
			}
			recKey, ok := s.records[section][guid]
			if !ok {
				return recKey, ErrRecordNotFound
			}
			return recKey, nil
		})

	if err != nil {
		return res.(Key), err
	}
	return res.(Key), nil
}

func (s *stash) recordAddSingleFlight(section SectionIdType, recGuid string, recKey Key) {
	key := string(section)
	_, _, _ = s.sfRecords.Do(key,
		func() (interface{}, error) {
			if s.records[section] == nil {
				// todo: move init and change single flight key
				s.records[section] = s.recordsInSection(section)
			}
			s.records[section][recGuid] = recKey
			return true, nil
		})
}

func (s *stash) recordRemoveSingleFlight(section SectionIdType, guid string) (Key, error) {
	key := string(section)
	res, err, _ := s.sfRecords.Do(key,
		func() (interface{}, error) {
			if s.records[section] == nil {
				// todo: move init and change single flight key
				s.records[section] = s.recordsInSection(section)
			}
			recKey, ok := s.records[section][guid]
			if !ok {
				return recKey, ErrRecordNotFound
			}
			delete(s.records[section], guid)
			return recKey, nil
		})

	if err != nil {
		return res.(Key), err
	}
	return res.(Key), nil
}

func (s *stash) recordsInSection(section SectionIdType) map[string]Key {
	res := make(map[string]Key)

	node, found := s.findRecord(section, counterRecordId)
	if !found {
		return res
	}

	it := s.iteratorAt(node)
	for it.pos == onmyway {
		if it.node.key.Section() != section {
			break
		}
		if it.node.key.Field() != headerFieldId {
			it.next()
			continue
		}
		header, err := s.getRecordHeader(node.key)
		if err != nil {
			s.sugar.Errorw("recordsInSection", "error", err)
			it.next()
			continue
		}
		if header.deleted {
			it.next()
			continue
		}

		res[header.guid] = it.node.key
		it.next()
	}

	return res
}

// Insert data
func (s *stash) Insert(section SectionIdType, data map[string]any) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	recId := s.newId(section)
	key := NewKey(section, recId, headerFieldId)
	header := newRecordHeader()
	s.m.Store(key, header)
	s.recordAddSingleFlight(section, header.guid, key)
	s.put(key)
	s.sugar.Debugw("insert header", "guid", header.guid, "key", key)

	for name, value := range data {
		fid := s.fieldIdSingleFlight(section, name)
		key := NewKey(section, recId, fid)
		s.m.Store(key, value)
		s.put(key)
		s.sugar.Debugw("insert field", "name", name, "key", key)
	}

	return header.guid
}

// Get record data
func (s *stash) Get(section SectionIdType, guid string) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, err := s.recordKeySingleFlight(section, guid)
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
		name, err := s.fieldNameSingleFlight(section, it.node.key.Field())
		if err != nil {
			s.sugar.Debugw("fieldNameSingleFlight", "err", err)
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

func (s *stash) Remove(section SectionIdType, guid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, err := s.recordRemoveSingleFlight(section, guid)
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
