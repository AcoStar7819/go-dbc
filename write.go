package dbc

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// WriteDBC writes the File structure into writer, forming a correct header.
func WriteDBC(w io.Writer, f *File) error {
	// First, write the header
	if err := binary.Write(w, binary.LittleEndian, f.Header); err != nil {
		return fmt.Errorf("write header error: %w", err)
	}
	// Then raw records
	if _, err := w.Write(f.RawRecords); err != nil {
		return fmt.Errorf("write records error: %w", err)
	}
	// And string block
	if _, err := w.Write(f.StringBlock); err != nil {
		return fmt.Errorf("write string block error: %w", err)
	}
	return nil
}

// WriteFile is a convenient wrapper to write to a file by name.
func WriteFile(filename string, dbcFile *File) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return WriteDBC(f, dbcFile)
}
