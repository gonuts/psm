// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "errors"
import "strconv"

// ErrRange indicates that a value is out of range for the target type.
var ErrRange = errors.New("value out of range")

// ErrSyntax indicates that a value does not have the right syntax for the target type.
var ErrSyntax = errors.New("invalid syntax")

// A NumError records a failed conversion.
type NumError struct {
	Func string // the failing function (ParseBool, ParseInt, ParseUint, ParseFloat)
	Num  string // the input
	Err  error  // the reason the conversion failed (ErrRange, ErrSyntax)
}

func (e *NumError) Error() string {
	return "strconv." + e.Func + ": " + `parsing "` + e.Num + `": ` + e.Err.Error()
}

func syntaxError(fn, str string) *NumError {
	return &NumError{fn, str, ErrSyntax}
}

func rangeError(fn, str string) *NumError {
	return &NumError{fn, str, ErrRange}
}

const intSize = 32 << uint(^uint(0)>>63)

const IntSize = intSize // number of bits in int, uint (32 or 64)

// Return the first number n such that n*base >= 1<<64.
func cutoff64(base int) uint64 {
	if base < 2 {
		return 0
	}
	return (1<<64-1)/uint64(base) + 1
}

// ParseUint is like ParseInt but for unsigned numbers.
func ParseUint(s []byte, base int, bitSize int) (n uint64, err error) {
	var cutoff, maxVal uint64

	if bitSize == 0 {
		bitSize = int(IntSize)
	}

	s0 := s
	switch {
	case len(s) < 1:
		err = ErrSyntax
		goto Error

	case 2 <= base && base <= 36:
		// valid base; nothing to do

	case base == 0:
		// Look for octal, hex prefix.
		switch {
		case s[0] == '0' && len(s) > 1 && (s[1] == 'x' || s[1] == 'X'):
			base = 16
			s = s[2:]
			if len(s) < 1 {
				err = ErrSyntax
				goto Error
			}
		case s[0] == '0':
			base = 8
		default:
			base = 10
		}

	default:
		err = errors.New("invalid base " + strconv.Itoa(base))
		goto Error
	}

	n = 0
	cutoff = cutoff64(base)
	maxVal = 1<<uint(bitSize) - 1

	for i := 0; i < len(s); i++ {
		var v byte
		d := s[i]
		switch {
		case '0' <= d && d <= '9':
			v = d - '0'
		case 'a' <= d && d <= 'z':
			v = d - 'a' + 10
		case 'A' <= d && d <= 'Z':
			v = d - 'A' + 10
		default:
			n = 0
			err = ErrSyntax
			goto Error
		}
		if int(v) >= base {
			n = 0
			err = ErrSyntax
			goto Error
		}

		if n >= cutoff {
			// n*base overflows
			n = 1<<64 - 1
			err = ErrRange
			goto Error
		}
		n *= uint64(base)

		n1 := n + uint64(v)
		if n1 < n || n1 > maxVal {
			// n+v overflows
			n = 1<<64 - 1
			err = ErrRange
			goto Error
		}
		n = n1
	}

	return n, nil

Error:
	return n, &NumError{"ParseUint", string(s0), err}
}
