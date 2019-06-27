package httproutegen

import (
	"errors"
	"fmt"
	"strings"
)

// SequencePart represent sequence part in component
type SequencePart struct {
	ByteMap           ByteMapper `json:"byte_map"`
	VariableName      string     `json:"variable_name"`
	VariableType      string     `json:"variable_type"`
	Converter         string     `json:"converter"`
	AliasVariableName []string   `json:"variable_name_aliases,omitempty"`
}

func (p *SequencePart) setSeqence(c []byte) (int, error) {
	progress := 0
	escapeMode := false
	ignoreBefore := -1
	var textBuf []byte
	for idx, ch := range c {
		if idx == 0 {
			if ch != '{' {
				return 0, errors.New("sequence must start with `{` character")
			}
			continue
		} else if idx <= ignoreBefore {
			continue
		}
		switch progress {
		case 0:
			ignoreBefore = p.ByteMap.SetByteMap(c[idx:], ',') + idx
			textBuf = make([]byte, 0)
			progress = 1
		case 1:
			if ch == ' ' {
				if len(textBuf) > 0 {
					p.VariableName = string(textBuf)
					textBuf = make([]byte, 0)
					progress = 2
				}
			} else {
				textBuf = append(textBuf, ch)
			}
		case 2:
			if ch == ',' || ch == '}' {
				if len(textBuf) <= 0 {
					return 0, fmt.Errorf("cannot have type for %v (index-in-group: %d)", p.VariableName, idx)
				}
				p.VariableType = string(textBuf)
				textBuf = make([]byte, 0)
				progress = 3
				if ch == '}' {
					return idx + 1, nil
				}
			} else if ch == ' ' {
				if len(textBuf) > 0 {
					return 0, fmt.Errorf("unexpected space for %v (index-in-group: %d)", p.VariableName, idx)
				}
			} else {
				textBuf = append(textBuf, ch)
			}
		case 3:
			if !escapeMode {
				if ch == '}' {
					p.Converter = strings.TrimSpace(string(textBuf))
					return idx + 1, nil
				}
				if ch == '\\' {
					escapeMode = true
					continue
				}
			} else {
				escapeMode = false
			}
			textBuf = append(textBuf, ch)
		}
	}
	return 0, fmt.Errorf("group not close: %v", p.VariableName)
}

// Equal check if two instance of SequencePart is equivalent.
func (p *SequencePart) Equal(other *SequencePart) bool {
	if (p.ByteMap != other.ByteMap) ||
		(p.Converter != other.Converter) ||
		(p.VariableType != other.VariableType) {
		return false
	}
	return true
}

// AttachVariableName add variable name to this part.
func (p *SequencePart) AttachVariableName(n string) {
	if (p.VariableName == n) || (p.VariableName == "") {
		p.VariableName = n
	} else {
		for _, aliasName := range p.AliasVariableName {
			if aliasName == n {
				return
			}
		}
		p.AliasVariableName = append(p.AliasVariableName, n)
	}
}
