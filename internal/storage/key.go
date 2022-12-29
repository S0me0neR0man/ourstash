package storage

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	keyLength = 16

	KeyLessThan int = -1
	KeyEqual    int = 0
	KeyMoreThan int = 1
)

type SectionIdType byte
type RecordIdType uint64
type FieldIdType uint16

// Key the synthetic unique key. All digit stored in BigEndian notation.
//
// [0] the section
//
// [1:9] the record id uint64
//
// [09:11] the field id uint16
//
// [11:16] reserved
type Key [keyLength]byte

// NewKey make new block key
func NewKey(sec SectionIdType, rec RecordIdType, field FieldIdType) Key {
	var k [keyLength]byte
	k[0] = byte(sec)
	binary.BigEndian.PutUint64(k[1:9], uint64(rec))
	binary.BigEndian.PutUint16(k[9:11], uint16(field))
	return k
}

func NewKeyFromBytes(b []byte) (Key, error) {
	var k [keyLength]byte
	if len(b) != keyLength {
		return k, errors.New("wrong length")
	}
	return k, nil
}

// String is Stringer implementation
func (k Key) String() string {
	return fmt.Sprintf("%s %s %s",
		hex.EncodeToString(k[0:1]),
		hex.EncodeToString(k[1:9]),
		hex.EncodeToString(k[9:11]),
	)
}

func (k Key) Section() SectionIdType {
	return SectionIdType(k[0])
}

func (k Key) Record() RecordIdType {
	return RecordIdType(binary.BigEndian.Uint64(k[1:9]))
}

func (k Key) Field() FieldIdType {
	return FieldIdType(binary.BigEndian.Uint16(k[9:11]))
}

func (k Key) Compare(other Key) int {
	if k.Section() < other.Section() {
		return KeyLessThan
	}
	if k.Section() > other.Section() {
		return KeyMoreThan
	}
	if k.Record() < other.Record() {
		return KeyLessThan
	}
	if k.Record() > other.Record() {
		return KeyMoreThan
	}
	if k.Field() < other.Field() {
		return KeyLessThan
	}
	if k.Field() > other.Field() {
		return KeyMoreThan
	}
	return KeyEqual
}
