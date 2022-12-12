package storage

import (
	"log"
)

type AuthUnit struct {
}

func NewAuthUnit() (*AuthUnit, error) {
	return &AuthUnit{}, nil
}

func (a *AuthUnit) String() string {
	return "AuthUnit"
}

func (a *AuthUnit) PutMiddleware(next PutHandler) PutHandler {
	return PutHandlerFunc(func(store Storager) error {
		log.Println("auth unit put handler")

		if next == nil {
			return nil
		}
		return next.Put(store)
	})
}
