package storage

import (
	"fmt"
	"time"
)

type SysUnit struct {
	user      string
	sectionId SectionIdType
	recordId  RecordIdType
	isDeleted bool
	lastMod   time.Time
}

func (s *SysUnit) PutMiddleware(handler PutHandler) PutHandler {
	//TODO implement me
	panic("implement me")
}

func NewSysUnit() (*SysUnit, error) {
	return &SysUnit{}, nil
}

func (s *SysUnit) String() string {
	return fmt.Sprintf("%s %d %d %v %v", s.user, s.sectionId, s.recordId, s.isDeleted, s.lastMod)
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

func (s *SysUnit) Put(storager Storager) error {
	//TODO implement me
	panic("implement me")
}
