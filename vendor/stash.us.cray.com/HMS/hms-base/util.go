// MIT License
//
// (C) Copyright [2019, 2021] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package base

import (
	"strings"
	"unicode"
)

// Returns true if s is alphanumeric only (only letters and numbers, no
// punctuation or spaces.
func IsAlphaNum(s string) bool {
	alphaNumFunc := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	// Make sure s is alphanumeric.
	idx := strings.IndexFunc(s, alphaNumFunc)
	if idx != -1 {
		return false
	}
	return true
}

// Returns true if s is numeric only (only numbers, no letters,
// punctuation or spaces.
func IsNumeric(s string) bool {
	numFunc := func(c rune) bool {
		return !unicode.IsNumber(c)
	}
	// Make sure s is numeric.
	idx := strings.IndexFunc(s, numFunc)
	if idx != -1 {
		return false
	}
	return true
}

// Remove leading zeros, i.e. for each run of numbers, trim off leading
// zeros so each run starts with either non-zero, or is a single zero.
func RemoveLeadingZeros(s string) string {
	//var b strings.Builder // Go 1.10
	b := []byte("")

	// base case
	length := len(s)
	if length < 2 {
		return s
	}
	// Look for 0 after letter and before number. Skip these and
	// pretend the previous value was still a letter for the next
	// round, to catch multiple leading zeros.
	i := 0
	lastLetter := true
	for ; i < length-1; i++ {
		if s[i] == '0' && lastLetter == true {
			if unicode.IsNumber(rune(s[i+1])) {
				// leading zero
				continue
			}
		}
		if unicode.IsNumber(rune(s[i])) {
			lastLetter = false
		} else {
			lastLetter = true
		}
		// b.WriteByte(s[i]) // Go 1.10
		b = append(b, s[i])
	}
	//b.WriteByte(s[i]) // Go 1.10
	//return b.String()
	b = append(b, s[i])
	return string(b)
}
