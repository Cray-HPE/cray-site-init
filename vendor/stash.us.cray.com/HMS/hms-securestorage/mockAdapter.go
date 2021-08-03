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

package securestorage

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
)

type InputStore struct {
	Key   string
	Value interface{}
}

type OutputStore struct {
	Err error
}

type MockStore struct {
	Input  InputStore
	Output OutputStore
}

type InputLookup struct {
	Key string
}

type OutputLookup struct {
	Output interface{}
	Err    error
}

type MockLookup struct {
	Input  InputLookup
	Output OutputLookup
}

type InputDelete struct {
	Key string
}

type OutputDelete struct {
	Err error
}

type MockDelete struct {
	Input  InputDelete
	Output OutputDelete
}

type InputLookupKeys struct {
	KeyPath string
}

type OutputLookupKeys struct {
	Klist []string
	Err   error
}

type MockLookupKeys struct {
	Input  InputLookupKeys
	Output OutputLookupKeys
}

type MockAdapter struct {
	StoreNum       int
	StoreData      []MockStore
	LookupNum      int
	LookupData     []MockLookup
	DeleteNum      int
	DeleteData     []MockDelete
	LookupKeysNum  int
	LookupKeysData []MockLookupKeys
}

func NewMockAdapter() (SecureStorage, *MockAdapter) {
	ss := &MockAdapter{}
	return ss, ss
}

func (ss *MockAdapter) Store(key string, value interface{}) error {
	i := ss.StoreNum
	if len(ss.StoreData) <= i {
		return fmt.Errorf("Unexpected call to MockStore")
	}
	ss.StoreNum++
	ss.StoreData[i].Input.Key = key
	ss.StoreData[i].Input.Value = value
	return ss.StoreData[i].Output.Err
}

func (ss *MockAdapter) Lookup(key string, output interface{}) error {
	if len(ss.LookupData) == 0 {
		return fmt.Errorf("Unexpected call to MockLookup: no data")
	}

	var i int

	if ss.LookupNum > -1 {
		i = ss.LookupNum
		if len(ss.LookupData) < (i + 1) {
			return fmt.Errorf("Unexpected call to MockLookup: less data than index")
		}
		ss.LookupNum++
		ss.LookupData[i].Input.Key = key
	} else {
		for i = 0; i < len(ss.LookupData); i++ {
			if key == ss.LookupData[i].Input.Key {
				break
			}
		}
		if i >= len(ss.LookupData) {
			return fmt.Errorf("Unexpected call to MockLookup: key not found")
		}
	}

	err := mapstructure.Decode(ss.LookupData[i].Output.Output, output)
	if err != nil {
		return err
	}
	return ss.LookupData[i].Output.Err
}

func (ss *MockAdapter) Delete(key string) error {
	i := ss.DeleteNum
	if len(ss.DeleteData) <= i {
		return fmt.Errorf("Unexpected call to MockDelete")
	}
	ss.DeleteNum++
	ss.DeleteData[i].Input.Key = key
	return ss.DeleteData[i].Output.Err
}

func (ss *MockAdapter) LookupKeys(keyPath string) ([]string, error) {
	if len(ss.LookupKeysData) == 0 {
		return nil, fmt.Errorf("Unexpected call to MockLookupKeys: no LookupKeysData")
	}

	var i int

	if ss.LookupKeysNum > -1 {
		i = ss.LookupKeysNum
		if len(ss.LookupKeysData) < (i + 1) {
			return nil, fmt.Errorf("Unexpected call to MockLookupKeys: less data than index")
		}
		ss.LookupKeysNum++
		ss.LookupKeysData[i].Input.KeyPath = keyPath
	} else {
		for i = 0; i < len(ss.LookupKeysData); i++ {
			if keyPath == ss.LookupKeysData[i].Input.KeyPath {
				break
			}
		}
		if i >= len(ss.LookupKeysData) {
			return nil, fmt.Errorf("Unexpected call to MockLookupKeys: keyPath not found")
		}
	}

	return ss.LookupKeysData[i].Output.Klist, ss.LookupKeysData[i].Output.Err
}
