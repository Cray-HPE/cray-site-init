package kubernetes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	attemps := 0
	testCases := []struct {
		Name     string
		Attempts int
		Sleep    string
		Function func(string) error
	}{
		{
			Name:     "should retry 5 times",
			Attempts: 5,
			Sleep:    "1ms",
			Function: func(test string) error {
				attemps += 1
				if attemps != 5 {
					return fmt.Errorf("always fail")
				}
				return nil
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			attemps = 0
			err := retry(tc.Attempts, tc.Sleep, tc.Function("asdf"))
			assert.True(t, err)
		})
	}
}
