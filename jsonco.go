// Package jsonco implements an io.Reader wrapper for JSON with C-style comments. It also has a utility method
// which translates byte offset into line and character positions for easier error reporting.
// The package is aware of multibyte characters.
package jsonco

import (
	"bytes"
	"errors"
	"io"
)

type state struct {
	// Data source.
	src io.Reader
	// Content of data source read into memory.
	bytes []byte
	// Wrapper for bytes.
	br *bytes.Reader
}

type ReadOffsetter interface {
	io.Reader
	LineAndChar(offset int64) (ln int, cn int, err error)
}

// New returns an io.Reader acting as proxy to r.
func New(r io.Reader) ReadOffsetter {
	return &state{src: r}
}

// Read reads bytes from the underlying reader replacing C-style comments and trailing commas with spaces.
func (st *state) Read(p []byte) (n int, err error) {
	if st.br == nil {
		if st.bytes, err = processInput(st.src); err != nil {
			return
		}
		st.br = bytes.NewReader(st.bytes)
	}
	return st.br.Read(p)
}

// LineAndChar calculates line and character position from the byte offset into underlying stream.
func (st *state) LineAndChar(offset int64) (int, int, error) {
	if offset < 0 {
		return -1, -1, errors.New("offset value cannot be negative")
	}

	br := bytes.NewReader(st.bytes)

	// Count lines and characters.
	lnum := 1
	cnum := 0
	// Number of bytes consumed.
	var count int64
	for {
		ch, size, err := br.ReadRune()
		if err == io.EOF {
			return -1, -1, errors.New("offset value too large")
		}
		count += int64(size)

		if ch == '\n' {
			lnum++
			cnum = 0
		} else {
			cnum++
		}
		//log.Println(offset, ch, string(ch), size, count, lnum, cnum)
		if count >= offset {
			break
		}
	}

	return lnum, cnum, nil
}

func isNL(c byte) bool {
	return c == '\n' || c == '\r'
}

func isWS(c byte) bool {
	return c == ' ' || c == '\t' || isNL(c)
}

func consumeComment(s []byte, i int) int {
	if i < len(s) && s[i] == '/' {
		s[i-1] = ' '
		for ; i < len(s) && !isNL(s[i]); i += 1 {
			s[i] = ' '
		}
	}
	if i < len(s) && s[i] == '*' {
		s[i-1] = ' '
		s[i] = ' '
		for ; i < len(s); i += 1 {
			if s[i] != '*' {
				// Do not remove new lines inside multiline comments.
				if s[i] != '\n' {
					s[i] = ' '
				}
			} else {
				s[i] = ' '
				i++
				if i < len(s) {
					if s[i] == '/' {
						s[i] = ' '
						break
					}
				}
			}
		}
	}
	return i
}

func processInput(r io.Reader) ([]byte, error) {
	buff := &bytes.Buffer{}
	_, err := io.Copy(buff, r)
	if err != nil {
		return nil, err
	}
	s := buff.Bytes()

	i := 0
	for i < len(s) {
		switch s[i] {
		case '"':
			i += 1
			for i < len(s) {
				if s[i] == '"' {
					i += 1
					break
				} else if s[i] == '\\' {
					i += 1
				}
				i += 1
			}
		case '/':
			i = consumeComment(s, i+1)
		case ',':
			j := i
			for {
				i += 1
				if i >= len(s) {
					break
				} else if s[i] == '}' || s[i] == ']' {
					s[j] = ' '
					break
				} else if s[i] == '/' {
					i = consumeComment(s, i+1)
				} else if !isWS(s[i]) {
					break
				}
			}
		default:
			i += 1
		}
	}

	return s, nil
}
