package storage

import "time"

type SysUnit struct {
	user      string
	sectionId SectionIdType
	recordId  RecordIdType
	isDeleted bool
	lastMod   time.Time
}

func NewSysUnit() (*SysUnit, error) {
	return &SysUnit{}, nil
}

func (s *SysUnit) User() string {
	return s.user
}

func (s *SysUnit) SectionId() SectionIdType {
	return s.sectionId
}

func (s *SysUnit) RecordId() RecordIdType {
	return s.recordId
}

func (s *SysUnit) IsDeleted() bool {
	return s.isDeleted
}

func (s *SysUnit) LastMod() time.Time {
	return s.lastMod
}

func (s *SysUnit) Name() string {
	return sysUnitName
}

func (s *SysUnit) Middleware(handler PutHandler) PutHandler {
	//TODO implement me
	panic("implement me")
}

func (s *SysUnit) PutHandler(next PutHandler) PutHandler {
	return PutHandlerFunc(func(k *Key) {
		if next == nil {
			return
		}
		next.Put(k)
	})
}
