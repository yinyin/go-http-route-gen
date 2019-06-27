package httproutegen

import (
	"errors"
	"log"
)

// SymbolType represent type of symbol.
type SymbolType int

// Symbol types
const (
	SymbolTypeNoop SymbolType = iota
	SymbolTypeByte
	SymbolTypeSequence
)

// Symbol represent one input byte or variable.
type Symbol struct {
	Type SymbolType `json:"symbol_type"`

	ByteValue byte `json:"byte_value,omitempty"`

	SequenceValue *SequencePart `json:"sequence_value,omitempty"`
	SequenceIndex int           `json:"sequence_index,omitempty"`
}

// NoopSymbol is an instance of NOOP symbol.
var NoopSymbol = Symbol{
	Type: SymbolTypeNoop,
}

func newByteSymbol(b byte) Symbol {
	return Symbol{
		Type:      SymbolTypeByte,
		ByteValue: b,
	}
}

func newSequenceSymbol(value *SequencePart, index int) Symbol {
	return Symbol{
		Type:          SymbolTypeSequence,
		SequenceValue: value,
		SequenceIndex: index,
	}
}

// SymbolScope represent one shared space of symbol parsing operation.
type SymbolScope struct {
	FoundSequences []*SequencePart `json:"found_sequences"`
}

func (scope *SymbolScope) attachSequencePart(seqPart *SequencePart) (int, *SequencePart) {
	for idx, part := range scope.FoundSequences {
		if part.Equal(seqPart) {
			part.AttachVariableName(seqPart.VariableName)
			return idx, part
		}
	}
	scope.FoundSequences = append(scope.FoundSequences, seqPart)
	return len(scope.FoundSequences) - 1, seqPart
}

// ParseComponent parse given bytes as component.
func (scope *SymbolScope) ParseComponent(c []byte) (result []Symbol, err error) {
	for len(c) > 0 {
		if ch := c[0]; ch == '{' {
			seqPart := &SequencePart{}
			nextIdx, err := seqPart.setSeqence(c)
			if nil != err {
				log.Printf("ERROR: failed on set sequence to part: %v", string(c))
				return nil, err
			}
			seqIndex, seqPart := scope.attachSequencePart(seqPart)
			result = append(result, newSequenceSymbol(seqPart, seqIndex))
			if nextIdx < len(c) {
				c = c[nextIdx:]
			} else {
				c = nil
			}
		} else if ch == '\\' {
			if len(c) < 2 {
				err = errors.New("escape at end of component")
				return
			}
			ch = c[1]
			result = append(result, newByteSymbol(ch))
			c = c[2:]
		} else {
			result = append(result, newByteSymbol(ch))
			if len(c) > 1 {
				c = c[1:]
			} else {
				c = nil
			}
		}
	}
	return
}
