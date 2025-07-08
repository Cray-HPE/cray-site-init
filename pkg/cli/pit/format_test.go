/*
MIT License

(C) Copyright 2022-2025 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/

package pit

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func Test_FormatCommand(t *testing.T) {
	b := bytes.NewBufferString("")
	mockWriteFunc := WriteLiveCDFunc(
		// create a mock function that will write to the buffer since we do not want to actually format anything
		func(device string, iso string, size string) error {
			fmt.Fprint(
				b,
				"write-script flag not set",
			)
			return nil
		},
	)
	cmd := formatCommand(mockWriteFunc)                 // create a new 'pit format' command using the mock function
	cmd.SetOut(b)                                       // replace the stdout with something that we can read programmatically
	cmd.SetErr(b)                                       // replace the stderr with something that we can read programmatically
	cmd.SetArgs([]string{"/dev/mock", "mock.iso", "1"}) // set the required args for the command
	// the write-script flag should be optional, so do not set it

	err := cmd.Execute() // execute the command
	if err != nil {
		t.Fatalf(
			"cmd.Execute() failed: %v",
			err,
		)
	}
	out, err := io.ReadAll(b) // read the output
	if err != nil {
		t.Fatal(err)
	}

	// if
	if string(out) != "write-script flag not set" {
		t.Fatalf(
			"expected \"%s\" got \"%s\"",
			"write-script flag not set",
			string(out),
		)
	}
}
