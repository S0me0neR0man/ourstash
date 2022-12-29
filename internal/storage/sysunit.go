package storage

import (
	"fmt"
	"log"
	"time"
)

type SysUnit struct {
	user      string
	isDeleted bool
	lastMod   time.Time
}

func NewSysUnit() (*SysUnit, error) {
	return &SysUnit{}, nil
}

func (s *SysUnit) String() string {
	return fmt.Sprintf("%s %v %v", s.user, s.isDeleted, s.lastMod)
}

func (s *SysUnit) Name() string {
	return ""
}

func (s *SysUnit) PutMiddleware(next PutHandler) PutHandler {
	return PutHandlerFunc(func(store Storager) error {
		s.user = store.User()
		s.isDeleted = false
		s.lastMod = time.Now()
		log.Printf("sysunit put handler user='%s' deleted=%v %s\n",
			s.user, s.isDeleted, s.lastMod.Format(time.RFC822))

		if next == nil {
			return nil
		}
		return next.Put(store)
	})
}
