package storage

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const ()

type SectionIdType byte
type RecordIdType uint64
type UnitIdType uint16

// Key the synthetic unique key. All digit stored in BigEndian notation.
//
// [0] the section
//
// [1:9] the record id uint64
//
// [9:11] the unit id uint16
//
// [11:13] the field name id uint16
//
// [13:17] block extra data
//
// [17:21] reserved
//
// record id = 0 reserved in all section
//  Section RecordId UnitId FieldId
//  ------0 -------0 -----0 ------0 next record id for section 0
//  ------1 -------0 -----0 ------0 next record id for section 1

type Key [20]byte

type Iterator interface {
	Next() (*Key, error)
}

// NewBlockKey make new block key
func NewBlockKey(section SectionIdType, record RecordIdType, unit UnitIdType) *Key {
	k := new(Key)
	k[0] = byte(section)
	binary.BigEndian.PutUint64(k[1:9], uint64(record))
	binary.BigEndian.PutUint16(k[9:11], uint16(unit))
	return k
}

// String is Stringer implementation
func (k Key) String() string {
	return fmt.Sprintf("%s %s %s %s %s",
		hex.EncodeToString(k[0:1]),
		hex.EncodeToString(k[1:9]),
		hex.EncodeToString(k[9:11]),
		hex.EncodeToString(k[11:13]),
		hex.EncodeToString(k[13:17]),
	)
}
