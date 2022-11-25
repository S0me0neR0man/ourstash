package storage

const ()

// Id the synthetic unique key
// Id[0] the section
// Id[1:9] the record id (binary.LittleEndian.PutUint64)
// Id[9:10] the block
// Id[10:12] field names (binary.LittleEndian.PutUint16)
// Id[12:16] field values (binary.LittleEndian.PutUint32)
type Id [15]byte

type Iterator interface {
	Next() (Id, error)
}

type Idier interface {
	Records(section byte) Iterator
	Blocks(section byte, record [8]byte) Iterator
	Fields(section byte, record [8]byte, block byte) Iterator
	Values(section byte, record [8]byte, block byte, field [2]byte) Iterator
}
