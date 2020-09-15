// (c) 2013 Rick Arnold. Licensed under the BSD license (see LICENSE).

package props

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf16"
)

const hexChars = "0123456789ABCDEFabcdef"

// scanner is used to parse the property file format as defined by Java. The
// parsing logic is based on the "Lexical Scanning in Go" presentation by
// Rob Pike at http://cuddle.googlecode.com/hg/talk/lex.html
type scanner struct {
	p *Properties

	// the key and value for the current line
	key   bytes.Buffer
	value bytes.Buffer

	// the current output buffer; either key or value
	current *bytes.Buffer

	// the current UTF-16 escapes (multiple escapes in a row are used to
	// represent characters that require more than 2 bytes)
	utfUnits []bytes.Buffer
}

func (s *scanner) finishEscape() stateFunc {
	s.utfUnits = nil
	if s.current == &s.key {
		return stateKey
	} else {
		return stateValue
	}
}

func (s *scanner) startUtfEscape() stateFunc {
	if s.utfUnits == nil {
		s.utfUnits = make([]bytes.Buffer, 0, 4)
	}
	s.utfUnits = append(s.utfUnits, bytes.Buffer{})
	return stateUtfEscape
}

func (s *scanner) finishUtfEscape() {

	if len(s.utfUnits) <= 0 {
		return
	}

	units := make([]uint16, 0, len(s.utfUnits))
	for _, v := range s.utfUnits {
		var unit uint16
		n, _ := fmt.Sscanf(strings.ToLower(v.String()), "%x", &unit)
		if n == 0 {
			s.current.WriteRune(unicode.ReplacementChar)
			s.finishEscape()
			return
		}
		units = append(units, unit)
	}

	for _, r := range utf16.Decode(units) {
		s.current.WriteRune(r)
	}
	s.finishEscape()
	return
}

func (s *scanner) checkEscape(ch rune) stateFunc {
	if ch == '\\' {
		if s.current == nil {
			s.current = &s.key
		}
		return stateEscape
	}
	s.finishUtfEscape()
	return nil
}

func (s *scanner) done() {
	if s.key.Len() > 0 {
		s.p.values[s.key.String()] = s.value.String()
	}
}

// stateFunc represents a single state in the scanner's state machine.
type stateFunc func(*scanner, rune) stateFunc

// stateNone is the default state at the beginning of each line.
func stateNone(s *scanner, ch rune) stateFunc {

	if next := s.checkEscape(ch); next != nil {
		return next
	}

	if ch == '#' || ch == '!' {
		return stateComment
	}

	if isWhitespace(ch) {
		return stateNone
	}

	s.current = &s.key
	s.current.WriteRune(ch)
	return stateKey
}

// stateComment indicates that the current line is a comment; all characters
// up to the next newline will be ignored.
func stateComment(s *scanner, ch rune) stateFunc {
	if ch == '\r' || ch == '\n' {
		return stateNone
	}
	return stateComment
}

// stateKey indicates that the key is being read; all characters up to the
// first (unescaped) whitespace, '=', or ':' will be considered part of the
// key.
func stateKey(s *scanner, ch rune) stateFunc {
	if next := s.checkEscape(ch); next != nil {
		return next
	}

	if ch == '=' || ch == ':' {
		s.current = &s.value
		return stateSeparatorChar
	}

	if ch == '\r' || ch == '\n' {
		return finishEntry(s)
	}

	if isWhitespace(ch) {
		s.current = &s.value
		return stateSeparator
	}

	s.current.WriteRune(ch)
	return stateKey
}

// stateSeparator indicates that whitespace between the key and value is being
// read.
func stateSeparator(s *scanner, ch rune) stateFunc {
	if next := s.checkEscape(ch); next != nil {
		return next
	}

	if ch == '=' || ch == ':' {
		return stateSeparatorChar
	}

	if ch == '\r' || ch == '\n' {
		return finishEntry(s)
	}

	if isWhitespace(ch) {
		return stateSeparator
	}

	s.current.WriteRune(ch)
	return stateValue
}

// stateSeparatorChar indicates that the '=' or ':' character or whitespace
// before the value is being read.
func stateSeparatorChar(s *scanner, ch rune) stateFunc {
	if next := s.checkEscape(ch); next != nil {
		return next
	}

	if ch == '\r' || ch == '\n' {
		return finishEntry(s)
	}

	if isWhitespace(ch) {
		return stateSeparatorChar
	}

	s.current.WriteRune(ch)
	return stateValue
}

// stateValue indicates that the value text is being read.
func stateValue(s *scanner, ch rune) stateFunc {
	if next := s.checkEscape(ch); next != nil {
		return next
	}

	if ch == '\r' || ch == '\n' {
		return finishEntry(s)
	}

	s.current.WriteRune(ch)
	return stateValue
}

// stateContinued indicates that an escaped newline or corresponding leading
// whitespace on the next line is being read. The first non-whitespace
// character will continue the key or value previously being read.
func stateContinued(s *scanner, ch rune) stateFunc {
	if isWhitespace(ch) {
		return stateContinued
	}

	if next := s.checkEscape(ch); next != nil {
		return next
	}

	s.current.WriteRune(ch)
	return s.finishEscape()
}

// stateEscape indicates that an escaped character is being read. Valid escapes
// will be replaced by special characters such as '\n'; invalid escapes will
// write the escaped character unchanged. Once the escaped character is read,
// normal scanning of the key or value resumes.
func stateEscape(s *scanner, ch rune) stateFunc {

	if ch == 'u' {
		return s.startUtfEscape()
	}

	s.finishUtfEscape()

	if ch == '\n' || ch == '\r' {
		return stateContinued
	}

	if ch == 't' {
		s.current.WriteRune('\t')
	} else if ch == 'n' {
		s.current.WriteRune('\n')
	} else if ch == 'r' {
		s.current.WriteRune('\r')
	} else if ch == 'f' {
		s.current.WriteRune('\f')
	} else {
		s.current.WriteRune(ch)
	}

	return s.finishEscape()
}

// stateUtfEscape indicates that a UTF-16 escape is being read. If the escape
// contains 4 hex digits it is added to the current list of escaped code units;
// otherwise the Unicode replacement character is used. Characters that require
// more than 2 bytes to represent will require multiple escapes in a row.
func stateUtfEscape(s *scanner, ch rune) stateFunc {

	if ch == 'u' {
		return stateUtfEscape
	}

	if strings.ContainsRune(hexChars, ch) {
		unit := &s.utfUnits[len(s.utfUnits)-1]
		unit.WriteRune(ch)
		if unit.Len() == 4 {
			if s.current == &s.key {
				return stateKey
			} else {
				return stateValue
			}
		} else {
			return stateUtfEscape
		}
	}

	s.current.WriteRune(unicode.ReplacementChar)
	return s.finishEscape()
}

// finishEntry handles the end of a property file entry and resets the
// scanner for the next entry
func finishEntry(s *scanner) stateFunc {
	s.p.values[s.key.String()] = s.value.String()
	s.key.Reset()
	s.value.Reset()
	s.current = &s.key
	return stateNone
}

// isWhitespace returns true for any character considered to be whitespace
// by the property file format.
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' || ch == '\f'
}
