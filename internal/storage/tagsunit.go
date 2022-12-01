package storage

import (
	"fmt"
	"log"
)

type TagsUnit struct {
	Tags []string
}

func NewTagsUnit(tags ...string) (*TagsUnit, error) {
	return &TagsUnit{Tags: tags}, nil
}

func (t *TagsUnit) Name() string {
	return "tags"
}

func (t *TagsUnit) String() string {
	return fmt.Sprintf("%v", t.Tags)
}

func (t *TagsUnit) PutHandle(next PutHandler) PutHandler {
	return PutHandlerFunc(func(store Storager) error {
		log.Println("tagsunit put handler")
		if next == nil {
			return nil
		}
		return next.Put(store)
	})
}

//func (t *TagsUnit) PutHandler(next PutHandler) PutHandler {
//	return PutHandlerFunc(func(store Storager) error {
//		log.Println("tagsunit put handler")
//		if next == nil {
//			return nil
//		}
//		return next.Put(store)
//	})
//}
