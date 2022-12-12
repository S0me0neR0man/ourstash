package storage

// todo: think about cyclic references

type storagerImpl struct {
	sectionId SectionIdType
	recId RecordIdType
	sm *SyncMapStorage
}

func newStoragerImpl(sId SectionIdType, rId RecordIdType, m *SyncMapStorage) *storagerImpl {
	return &storagerImpl{sectionId: sId, recId: rId, sm: m}
}

func (sc *storagerImpl) User() string {
	return "TODO implement me"
}

func (sc *storagerImpl) Put(key FieldIdType, data any) error {

	return nil
}