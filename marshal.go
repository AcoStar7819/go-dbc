package dbc

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
)

// UnmarshalRecords reads raw records into a slice of user-defined structures
// (e.g., *[]ChrClassesRecord). It uses tags like: `dbc:"column=0,type=uint32"`.
// Supports basic types: int32/uint32/float32/float64 and strings (via offset).
// For localized columns (offset + mask), the logic can be extended if needed.
func UnmarshalRecords(dbcFile *File, userStructSlice interface{}) error {
	rv := reflect.ValueOf(userStructSlice)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("UnmarshalRecords: userStructSlice must be pointer to a slice")
	}

	sliceVal := rv.Elem()
	sliceVal.SetLen(0) // clear the slice before filling

	recordCount := int(dbcFile.Header.RecordCount)
	recordSize := int(dbcFile.Header.RecordSize)

	// Determine the element type (structure)
	elemType := sliceVal.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("UnmarshalRecords: slice element must be a struct")
	}

	raw := dbcFile.RawRecords

	// Parse each record
	for i := 0; i < recordCount; i++ {
		recBytes := raw[i*recordSize : (i+1)*recordSize]
		// Create a new element (structure)
		newElem := reflect.New(elemType).Elem()

		// Iterate through structure fields
		for fIndex := 0; fIndex < newElem.NumField(); fIndex++ {
			fieldVal := newElem.Field(fIndex)
			fieldType := newElem.Type().Field(fIndex)

			tag := fieldType.Tag.Get("dbc")
			if tag == "" {
				// Skip field if tag is absent
				continue
			}

			// Parse the tag
			columnIndex, colType := parseTag(tag)

			// Prevent out-of-bound field access
			if columnIndex < 0 || columnIndex >= int(dbcFile.Header.FieldCount) {
				return fmt.Errorf("field %s has invalid column index %d", fieldType.Name, columnIndex)
			}
			// Byte offset within record
			offsetInRecord := 4 * columnIndex // each column = 4 bytes in classic DBC
			if offsetInRecord+4 > len(recBytes) {
				return fmt.Errorf("record %d, field %s out of bounds in DBC data", i, fieldType.Name)
			}

			// Read 4 bytes (or more) depending on colType
			switch strings.ToLower(colType) {
			case "int", "int32":
				v := binary.LittleEndian.Uint32(recBytes[offsetInRecord : offsetInRecord+4])
				fieldVal.SetInt(int64(int32(v)))

			case "uint32":
				v := binary.LittleEndian.Uint32(recBytes[offsetInRecord : offsetInRecord+4])
				fieldVal.SetUint(uint64(v))

			case "float", "float32":
				bits := binary.LittleEndian.Uint32(recBytes[offsetInRecord : offsetInRecord+4])
				floatVal := float32FromBits(bits)
				fieldVal.SetFloat(float64(floatVal))

			case "float64":
				return fmt.Errorf("float64 columns not supported by default 4-byte DBC columns")

			case "string":
				off := binary.LittleEndian.Uint32(recBytes[offsetInRecord : offsetInRecord+4])
				strVal := readCString(dbcFile.StringBlock, off)
				fieldVal.SetString(strVal)

			case "locstring":
				off := binary.LittleEndian.Uint32(recBytes[offsetInRecord : offsetInRecord+4])
				maskCol := columnIndex + 1
				if maskCol >= int(dbcFile.Header.FieldCount) {
					return fmt.Errorf("locstring at field %s: not enough columns for mask", fieldType.Name)
				}
				maskOffset := 4 * maskCol
				mask := binary.LittleEndian.Uint32(recBytes[maskOffset : maskOffset+4])

				if mask != 0 {
					return fmt.Errorf("locstring mask is not zero (mask=%d), field=%s", mask, fieldType.Name)
				}

				strVal := readCString(dbcFile.StringBlock, off)
				fieldVal.SetString(strVal)

			default:
				return fmt.Errorf("unsupported column type %q in field %s", colType, fieldType.Name)
			}
		}

		sliceVal = reflect.Append(sliceVal, newElem)
	}
	rv.Elem().Set(sliceVal)
	return nil
}

// MarshalRecords performs the reverse operation — forms raw records in a DBCFile
// from the given slice of structures.
// Simplified: generates a new stringBlock and sets string offsets.
// For localized strings with a bitmask (0), reserves 2 columns (offset, mask=0).
func MarshalRecords(userStructSlice interface{}) (*File, error) {
	rv := reflect.ValueOf(userStructSlice)
	if rv.Kind() != reflect.Slice {
		return nil, fmt.Errorf("MarshalRecords: userStructSlice must be a slice")
	}

	count := rv.Len()
	if count == 0 {
		return nil, fmt.Errorf("empty slice")
	}

	elemType := rv.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("MarshalRecords: slice element must be a struct")
	}

	maxColIndex := 0
	for i := 0; i < elemType.NumField(); i++ {
		tag := elemType.Field(i).Tag.Get("dbc")
		colIdx, _ := parseTag(tag)
		if colIdx > maxColIndex {
			maxColIndex = colIdx
		}
	}
	fieldCount := maxColIndex + 1
	recordSize := fieldCount * 4

	rawRecords := make([]byte, count*recordSize)

	strPool := make(map[string]uint32)
	var strBlock []byte

	strPool[""] = 0
	strBlock = append(strBlock, 0) // start with '\0'

	putString := func(s string) uint32 {
		if off, ok := strPool[s]; ok {
			return off
		}
		off := uint32(len(strBlock))
		strPool[s] = off
		strBlock = append(strBlock, []byte(s)...)
		strBlock = append(strBlock, 0)
		return off
	}

	for i := 0; i < count; i++ {
		elemVal := rv.Index(i)
		recOff := i * recordSize

		for fIndex := 0; fIndex < elemVal.NumField(); fIndex++ {
			fieldVal := elemVal.Field(fIndex)
			fieldType := elemType.Field(fIndex)
			tag := fieldType.Tag.Get("dbc")
			colIdx, colType := parseTag(tag)
			if colIdx < 0 {
				continue
			}

			byteOff := recOff + colIdx*4

			switch strings.ToLower(colType) {
			case "int", "int32":
				v := int32(fieldVal.Int())
				binary.LittleEndian.PutUint32(rawRecords[byteOff:byteOff+4], uint32(v))
			case "uint32":
				v := uint32(fieldVal.Uint())
				binary.LittleEndian.PutUint32(rawRecords[byteOff:byteOff+4], v)
			case "float", "float32":
				f := float32(fieldVal.Float())
				bits := float32ToBits(f)
				binary.LittleEndian.PutUint32(rawRecords[byteOff:byteOff+4], bits)
			case "string":
				s := fieldVal.String()
				off := putString(s)
				binary.LittleEndian.PutUint32(rawRecords[byteOff:byteOff+4], off)
			case "locstring":
				s := fieldVal.String()
				off := putString(s)
				binary.LittleEndian.PutUint32(rawRecords[byteOff:byteOff+4], off)

				if colIdx+1 >= fieldCount {
					return nil, fmt.Errorf("locstring field not enough space for mask")
				}
				byteOffMask := recOff + (colIdx+1)*4
				binary.LittleEndian.PutUint32(rawRecords[byteOffMask:byteOffMask+4], 0)
			default:
				return nil, fmt.Errorf("unsupported marshal type %q in field %s", colType, fieldType.Name)
			}
		}
	}

	hdr := Header{
		Magic:           MagicWDBC,
		RecordCount:     uint32(count),
		FieldCount:      uint32(fieldCount),
		RecordSize:      uint32(recordSize),
		StringBlockSize: uint32(len(strBlock)),
	}

	return &File{
		Header:      hdr,
		RawRecords:  rawRecords,
		StringBlock: strBlock,
	}, nil
}
