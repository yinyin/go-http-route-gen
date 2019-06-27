package httproutegen

import (
	"fmt"
	"log"
)

func computeBitMapParam(b byte) (mapIndex int, bitOffset uint) {
	bVal := uint(b)
	if 0 != (bVal & 64) {
		mapIndex = 1
	}
	bitOffset = bVal & 63
	return
}

// ByteMapper record how bytes map to scalar data type for handler arguments
type ByteMapper struct {
	bits [2]uint64
}

// MarshalJSON implements Marshaler interface of encoding/json package.
func (m *ByteMapper) MarshalJSON() ([]byte, error) {
	b0, b1 := m.ByteMap()
	r := fmt.Sprintf("\"0x%016X 0x%016X\"", b0, b1)
	return []byte(r), nil
}

// HasByte check if given byte is enabled in this mapper.
func (m *ByteMapper) HasByte(b byte) bool {
	if b > 127 {
		log.Printf("WARN: register byte > 127: %v", b)
		return false
	}
	bIndex, offset := computeBitMapParam(b)
	return ((m.bits[bIndex] & (1 << offset)) != 0)
}

func (m *ByteMapper) enableByte(b byte) {
	if b > 127 {
		log.Printf("WARN: register byte > 127: %v", b)
		return
	}
	bIndex, offset := computeBitMapParam(b)
	m.bits[bIndex] = m.bits[bIndex] | (1 << offset)
}

func (m *ByteMapper) enableByteRange(b0, b1 byte) {
	if (b0 > 127) || (b1 > 127) {
		log.Printf("WARN: register byte range > 127: %v - %v", b0, b1)
		return
	}
	if b1 < b0 {
		b0, b1 = b1, b0
	}
	for b := b0; b <= b1; b++ {
		m.enableByte(b)
	}
}

func (m *ByteMapper) enablePrintables() {
	m.enableByteRange(0x20, 0x7E)
}

func (m *ByteMapper) disableByte(b byte) {
	if b > 127 {
		log.Printf("WARN: register byte > 127: %v", b)
		return
	}
	bIndex, offset := computeBitMapParam(b)
	m.bits[bIndex] = m.bits[bIndex] & (^uint64(1 << offset))
}

func (m *ByteMapper) disableByteRange(b0, b1 byte) {
	if (b0 > 127) || (b1 > 127) {
		log.Printf("WARN: register byte range > 127: %v - %v", b0, b1)
		return
	}
	if b1 < b0 {
		b0, b1 = b1, b0
	}
	for b := b0; b <= b1; b++ {
		m.disableByte(b)
	}
}

// SetByteMap byte mask configuration in string form.
func (m *ByteMapper) SetByteMap(c []byte, stopByte byte) int {
	b := len(c)
	var buffered byte
	inverseMode := false
	escapeFlag := false
	rangingFlag := false
	for i := 0; i <= b; i++ {
		var ch byte
		if i < b {
			ch = c[i]
		}
		if (i == 0) && (ch == '^') {
			inverseMode = true
			m.enablePrintables()
			continue
		}
		if escapeFlag {
			escapeFlag = false
		} else if ch == stopByte {
			ch = 0
		} else if ch == '\\' {
			escapeFlag = true
			continue
		} else if ch == '-' {
			rangingFlag = true
			continue
		}
		if rangingFlag {
			if inverseMode {
				m.disableByteRange(buffered, ch)
			} else {
				m.enableByteRange(buffered, ch)
			}
			buffered = 0
			rangingFlag = false
			continue
		}
		if buffered != 0 {
			if inverseMode {
				m.disableByte(buffered)
			} else {
				m.enableByte(buffered)
			}
		}
		if ch == 0 {
			return i
		}
		buffered = ch
	}
	return b
}

// ByteMap return current bit mask of bytes enablement.
func (m *ByteMapper) ByteMap() (uint64, uint64) {
	return m.bits[0], m.bits[1]
}
