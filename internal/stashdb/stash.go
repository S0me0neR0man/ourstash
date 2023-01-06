package stashdb

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"ourstash/data"
)

type OperationType string
type GUIDType string

const (
	//	metadataSection  SectionIdType = 0
	//  usersRecord RecordIdType = 1

	metadataRecordId RecordIdType = 0
	counterFieldId   FieldIdType  = 0
	headerFieldId    FieldIdType  = 0

	InsertOperation OperationType = "insert"
	UpdateOperation OperationType = "update"

	fileName = "db/stash.data"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrFieldNotFound  = errors.New("field not found")
	ErrNotImplemented = errors.New("not implemented")
)

type recordHeader struct {
	Guid      GUIDType
	Next      RecordIdType
	Operation OperationType
	OpTime    time.Time
	Deleted   bool
}

func newRecordHeader(op OperationType) recordHeader {
	return recordHeader{
		Guid:      GUIDType(uuid.New().String()),
		OpTime:    time.Now(),
		Deleted:   false,
		Operation: op,
	}
}

type Record struct {
	guid GUIDType
	data map[string]any
}

// Stash the in-memory NoSQL key-value tread safe stashdb.
// Most important part - synthetic key (see Key type)
type Stash struct {
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

func NewStash(logger *zap.Logger) *Stash {
	return &Stash{
		redBlackTree: redBlackTree{},
		sugar:        logger.Sugar(),
		fields:       make(map[SectionIdType]map[string]FieldIdType, 0),
		records:      make(map[SectionIdType]map[GUIDType]Key, 0),
	}
}

func (s *Stash) newId(section SectionIdType) RecordIdType {
	key := NewKey(section, metadataRecordId, counterFieldId)
	var firstId uint64 = 1
	aid, loaded := s.m.LoadOrStore(key, &firstId)
	if !loaded {
		s.put(key) // todo: move to init section
		return RecordIdType(firstId)
	}
	return RecordIdType(atomic.AddUint64(aid.(*uint64), 1))
}

//func (s *Stash) findRecord(section SectionIdType, record RecordIdType, field FieldIdType) (*redBlackNode, bool) {
//	it := s.iterator()
//	for flag := it.Next(); flag; flag = it.Next() {
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

//func (s *Stash) getStringValue(key Key) (string, error) {
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

func (s *Stash) getRecordHeader(key Key) (recordHeader, error) {
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

func (s *Stash) fieldIdSFG(section SectionIdType, fieldName string) FieldIdType {
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

func (s *Stash) fieldNameSFG(section SectionIdType, fieldId FieldIdType) (string, error) {
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

func (s *Stash) recordKeySFG(section SectionIdType, guid GUIDType) (Key, error) {
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

func (s *Stash) recordAddSFG(section SectionIdType, guid GUIDType, recKey Key) {
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

func (s *Stash) recordRemoveSFG(section SectionIdType, guid GUIDType) (Key, error) {
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

func (s *Stash) putHeader(section SectionIdType, f func() recordHeader) (GUIDType, RecordIdType) {
	recId := s.newId(section)
	key := NewKey(section, recId, headerFieldId)
	header := f()
	s.m.Store(key, header)
	s.recordAddSFG(section, header.Guid, key)
	s.put(key)

	s.sugar.Debugw("put header", "Operation", header.Operation, "Guid", header.Guid, "key", key)
	return header.Guid, recId
}

func (s *Stash) putData(section SectionIdType, recId RecordIdType, data map[string]any) {
	for name, value := range data {
		fid := s.fieldIdSFG(section, name)
		key := NewKey(section, recId, fid)
		s.m.Store(key, value)
		s.put(key)
		s.sugar.Debugw("put data", "name", name, "key", key)
	}
}

// Insert data
func (s *Stash) Insert(section SectionIdType, data map[string]any) GUIDType {
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
func (s *Stash) Get(section SectionIdType, guid GUIDType) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, err := s.recordKeySFG(section, guid)
	if err != nil {
		return nil, err
	}
	s.sugar.Debugw("get", "Guid", guid, "key", key)

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
func (s *Stash) Remove(section SectionIdType, guid GUIDType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key, err := s.recordRemoveSFG(section, guid)
	if err != nil {
		return err
	}
	s.sugar.Debugw("remove", "Guid", guid, "key", key)

	var header recordHeader
	header, err = s.getRecordHeader(key)
	if err != nil {
		return err
	}
	header.Deleted = true
	s.m.Store(key, header)

	return nil
}

// Update data
func (s *Stash) Update(section SectionIdType, guid GUIDType, data map[string]any) error {
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
	prevHeader.Deleted = true
	prevHeader.OpTime = time.Now()

	guid, recId := s.putHeader(section, func() recordHeader {
		header := newRecordHeader(UpdateOperation)
		header.Guid = guid
		return header
	})
	s.putData(section, recId, data)

	prevHeader.Next = recId
	s.m.Store(prevKey, prevHeader)

	s.sugar.Debugw("update", "Guid", guid, "prevKey", prevKey)
	return nil
}

// Find data
func (s *Stash) Find(ctx context.Context, section SectionIdType, f func(*map[string]any) (bool, bool)) ([]Record, error) {
	var founded []Record

	s.recordsMu.Lock()
	if s.records[section] == nil {
		s.records[section] = make(map[GUIDType]Key) // todo: make all on start (init) and remove lock
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

		ok, stop := true, false
		if f != nil {
			ok, stop = f(&data)
		}

		if ok {
			founded = append(founded, Record{
				guid: guid,
				data: data,
			})
		}

		if stop {
			break
		}
	}
	return founded, nil
}

// SaveToDisk save data to disk
func (s *Stash) SaveToDisk(ctx context.Context) error {
	m := s.copyData(ctx)
	file, err := os.OpenFile(data.Path(fileName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	gob.Register(recordHeader{})
	err = gob.NewEncoder(file).Encode(m)
	if err != nil {
		return err
	}

	return nil
}

func (s *Stash) copyData(ctx context.Context) map[Key]any {
	ret := make(map[Key]any)

	s.mu.RLock()
	defer s.mu.RUnlock()

	s.m.Range(func(key, value any) bool {
		ret[key.(Key)] = value
		return true
	})

	return ret
}
