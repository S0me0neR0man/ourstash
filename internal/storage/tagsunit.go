package storage

import "fmt"

type TagsUnit struct {
	Tags []string
}

func NewTagsUnit(tags ...string) (*TagsUnit, error) {
	return &TagsUnit{Tags: tags}, nil
}

func (t *TagsUnit) Name() string {
	return "tags"
}

func (t *TagsUnit) PutHandler(next PutHandler) PutHandler {
	return PutHandlerFunc(func(k *Key) {
		next.Put(k)
	})
}

func (t *TagsUnit) String() string {
	return fmt.Sprintf("%v", t.Tags)
}
