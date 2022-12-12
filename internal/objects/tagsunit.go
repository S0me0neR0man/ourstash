package objects

import (
	"fmt"
	"log"

	"ourstash/internal/storage"
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

func (t *TagsUnit) PutMiddleware(next storage.PutHandler) storage.PutHandler {
	return storage.PutHandlerFunc(func(store storage.Storager) error {
		log.Println("tagsunit put handler")
		if next == nil {
			return nil
		}
		return next.Put(store)
	})
}

