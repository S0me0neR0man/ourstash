package storage

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

const (
	metadataSection  SectionIdType = 0
	metadataRecordId RecordIdType  = 0
	counterFieldId   FieldIdType   = 0
	headerFieldId    FieldIdType   = 0
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrFieldNotFound  = errors.New("field not found")
	ErrNotImplemented = errors.New("not implemented")
)

type recordHeader struct {
	guid    string
	next    RecordIdType
	time    time.Time
	deleted bool
}

func newRecordHeader() recordHeader {
	return recordHeader{
		guid:    uuid.New().String(),
		time:    time.Now(),
		deleted: false,
	}
}

// stash the in-memory NoSQL key-value tread safe storage.
// Most important part - synthetic key (see Key type)
type stash struct {
	redBlackTree
	m     sync.Map
	sugar *zap.SugaredLogger

	md map[SectionIdType]map[string]FieldIdType
	sf singleflight.Group
}

func newStash(logger *zap.Logger) *stash {
	return &stash{
		redBlackTree: redBlackTree{},
		sugar:        logger.Sugar(),
		md:           make(map[SectionIdType]map[string]FieldIdType, 0),
	}
}

func (s *stash) newId(section SectionIdType) RecordIdType {
	key := NewKey(section, metadataRecordId, counterFieldId)
	aid, loaded := s.m.LoadOrStore(key, RecordIdType(1))
	id := aid.(RecordIdType)
	if !loaded {
		s.put(key)
	}
	return RecordIdType(id)
}

// findRecord find first node in record
//
// IMPORTANT: does not provide thread safety
func (s *stash) findRecord(section SectionIdType, record RecordIdType) (*redBlackNode, bool) {
	it := s.iterator()
	for flag := it.next(); flag; it.next() {
		if it.node.key.Section() < section || it.node.key.Record() < record {
			continue
		}
		if it.node.key.Section() == section && it.node.key.Record() == record {
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
		return "", fmt.Errorf("%s value is not string (key %s)", msg, key.String())
	}

	return str, nil
}

// fieldId returns field id, register if new field passed
//
// thread safe
func (s *stash) fieldId(section SectionIdType, fieldName string) FieldIdType {
	key := string(section) + "-" + fieldName
	res, _, _ := s.sf.Do(key,
		func() (interface{}, error) {
			if s.md[section] == nil {
				// todo: move init
				s.md[section] = s.fields(section)
			}
			fid, ok := s.md[section][fieldName]
			if !ok {
				fid = FieldIdType(len(s.md[section]) + 1)
				key := NewKey(section, metadataRecordId, fid)
				s.m.Store(key, fieldName)
				s.put(key)
				s.md[section][fieldName] = fid
			}
			return fid, nil
		})

	return res.(FieldIdType)
}

// fields returns all fields in section, all errors are ignored
//
// IMPORTANT: does not provide thread safety
func (s *stash) fields(section SectionIdType) map[string]FieldIdType {
	res := make(map[string]FieldIdType)

	node, found := s.findRecord(section, metadataRecordId)
	if !found {
		return res
	}

	it := s.iteratorAt(node)
	for it.pos == onmyway {
		if it.node.key.Section() != section || it.node.key.Record() != metadataRecordId {
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

// fieldName returns field name
//
// thread safe
func (s *stash) fieldName(section SectionIdType, fieldId FieldIdType) (string, error) {
	key := string(section) + "-" + strconv.Itoa(int(fieldId))
	res, _, _ := s.sf.Do(key,
		func() (interface{}, error) {
			if s.md[section] == nil {
				// todo: move init
				s.md[section] = s.fields(section)
			}
			for fname, fid := range s.md[section] {
				if fid == fieldId {
					return fname, nil
				}
			}
			return 0, ErrFieldNotFound
		})

	return res.(string), nil
}

// Insert data in stash
func (s *stash) Insert(section SectionIdType, data map[string]any) RecordIdType {
	rec := s.newId(section)

	key := NewKey(section, rec, headerFieldId)
	s.m.Store(key, newRecordHeader())
	s.put(key)

	for name, value := range data {
		s.putField(section, rec, name, value)
	}

	return rec
}

func (s *stash) putField(section SectionIdType, record RecordIdType, fieldName string, fieldValue any) {
	fid := s.fieldId(section, fieldName)
	key := NewKey(section, record, fid)
	s.m.Store(key, fieldValue)
	s.put(key)
}

// Get record data
func (s *stash) Get(section SectionIdType, record RecordIdType) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	node, found := s.findRecord(section, record)
	if !found {
		return nil, ErrRecordNotFound
	}

	res := make(map[string]any)
	it := s.iteratorAt(node)
	for it.pos == onmyway {
		if it.node.key.Section() != section || it.node.key.Record() != record {
			break
		}
		if it.node.key.Field() == headerFieldId {
			it.next()
			continue
		}
		name, err := s.fieldName(section, it.node.key.Field())
		if err != nil {
			return nil, err
		}
		value, ok := s.m.Load(it.node.key)
		if !ok {
			return nil, errors.New("impossible, value stolen")
		}
		res[name] = value

		it.next()
	}

	return res, nil
}
