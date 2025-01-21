package dbc

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// ReadDBC reads the entire DBC file from reader into a File structure.
func ReadDBC(r io.Reader) (*File, error) {
	var hdr Header
	// Read the header
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, fmt.Errorf("read header error: %w", err)
	}

	if hdr.Magic != MagicWDBC {
		return nil, fmt.Errorf("not a WDBC file or bad magic: 0x%08X", hdr.Magic)
	}

	// Read raw records
	recordsSize := hdr.RecordCount * hdr.RecordSize
	rawRecords := make([]byte, recordsSize)
	if _, err := io.ReadFull(r, rawRecords); err != nil {
		return nil, fmt.Errorf("failed to read records: %w", err)
	}

	// Read string block
	stringBlock := make([]byte, hdr.StringBlockSize)
	if _, err := io.ReadFull(r, stringBlock); err != nil {
		return nil, fmt.Errorf("failed to read string block: %w", err)
	}

	return &File{
		Header:      hdr,
		RawRecords:  rawRecords,
		StringBlock: stringBlock,
	}, nil
}

// ReadFile is a convenient wrapper to read from a file by name.
func ReadFile(filename string) (*File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadDBC(f)
}
