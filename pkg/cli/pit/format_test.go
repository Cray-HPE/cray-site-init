package pit

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func Test_FormatCommand(t *testing.T) {
	b := bytes.NewBufferString("")
	mockWriteFunc := WriteLiveCDFunc( // create a mock function that will write to the buffer since we do not want to actually format anything
		func(device string, iso string, size string) error {
			fmt.Fprint(b, "write-script flag not set")
			return nil
		})
	cmd := formatCommand(mockWriteFunc)                 // create a new 'pit format' command using the mock function
	cmd.SetOut(b)                                       // replace the stdout with something that we can read programmatically
	cmd.SetErr(b)                                       // replace the stderr with something that we can read programmatically
	cmd.SetArgs([]string{"/dev/mock", "mock.iso", "1"}) // set the required args for the command
	// the write-script flag should be optional, so do not set it

	err := cmd.Execute() // execute the command
	if err != nil {
		t.Fatalf("cmd.Execute() failed: %v", err)
	}
	out, err := io.ReadAll(b) // read the output
	if err != nil {
		t.Fatal(err)
	}

	// if
	if string(out) != "write-script flag not set" {
		t.Fatalf("expected \"%s\" got \"%s\"", "write-script flag not set", string(out))
	}
}
