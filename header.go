package dbc

// MagicWDBC is the constant corresponding to "WDBC" in ASCII.
// 0x57 = 'W', 0x44 = 'D', 0x42 = 'B', 0x43 = 'C'
const MagicWDBC = 0x43424457

// Header describes the header of a DBC file.
type Header struct {
	Magic           uint32 // should always be MagicWDBC (0x43424457)
	RecordCount     uint32
	FieldCount      uint32
	RecordSize      uint32
	StringBlockSize uint32
}

// File describes the structure of a read DBC file in raw form.
type File struct {
	Header      Header
	RawRecords  []byte // length = RecordCount * RecordSize
	StringBlock []byte // length = StringBlockSize
}
