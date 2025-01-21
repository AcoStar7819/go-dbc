package dbc

import (
	"fmt"
	"math"
	"strings"
)

// parseTag parses a tag string like "column=0,type=uint32".
// Returns (index, typeName).
func parseTag(tag string) (int, string) {
	if tag == "" {
		return -1, ""
	}
	parts := strings.Split(tag, ",")
	var colIndex int
	var colType string
	for _, p := range parts {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
		switch k {
		case "column":
			fmt.Sscanf(v, "%d", &colIndex)
		case "type":
			colType = v
		}
	}
	return colIndex, colType
}

// readCString reads a C-style null-terminated string from stringBlock starting at offset.
func readCString(block []byte, offset uint32) string {
	if offset >= uint32(len(block)) {
		return ""
	}
	// Look for null byte
	b := block[offset:]
	idx := 0
	for idx < len(b) && b[idx] != 0 {
		idx++
	}
	return string(b[:idx])
}

// float32FromBits converts 32 bits (uint32) to float32.
func float32FromBits(bits uint32) float32 {
	return math.Float32frombits(bits)
}

// float32ToBits converts float32 to 32 bits (uint32).
func float32ToBits(f float32) uint32 {
	return math.Float32bits(f)
}
