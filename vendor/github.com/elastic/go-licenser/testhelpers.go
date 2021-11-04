// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

const fixtures = "fixtures"

func copyFixtures(t *testing.T, dest string) func() {
	if err := copy(fixtures, dest); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := os.RemoveAll(dest); err != nil {
			t.Fatal(err)
		}
	}
}

func copy(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return dcopy(src, dest, info)
	}
	return fcopy(src, dest, info)
}

func fcopy(src, dest string, info os.FileInfo) error {
	f, err := os.Create(
		strings.Replace(dest, ".testdata", ".go", 1),
	)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

func dcopy(src, dest string, info os.FileInfo) error {
	if err := os.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}

	infs, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for i := range infs {
		var source = filepath.Join(src, infs[i].Name())
		var destination = filepath.Join(dest, infs[i].Name())
		if err := copy(source, destination); err != nil {
			return err
		}
	}

	return nil
}

func hashDirectories(t *testing.T, src, dest string) {
	var srcHash = sha1.New()
	var dstHash = sha1.New()
	t.Logf("===== Walking %s =====\n", src)
	if err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil || path == src {
			return nil
		}

		t.Log(fmt.Sprint(info.Name(), " => ", info.Size()))
		io.WriteString(srcHash, fmt.Sprint(info.Name(), info.Size()))
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	t.Logf("===== Walking %s =====\n", dest)
	if err := filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
		if err != nil || path == dest {
			return nil
		}

		t.Log(fmt.Sprint(info.Name(), " => ", info.Size()))
		io.WriteString(dstHash, fmt.Sprint(info.Name(), info.Size()))
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	t.Log("===========================")
	var srcSum, dstSum = srcHash.Sum(nil), dstHash.Sum(nil)
	if bytes.Compare(srcSum, dstSum) != 0 {
		t.Errorf("Contents of %s are not the same as %s", src, dest)
		t.Errorf("src folder hash: %x", srcSum)
		t.Errorf("dst folder hash: %x", dstSum)
	}
}

func goosPathError(code int, p string) error {
	var opName = "stat"
	if runtime.GOOS == "windows" {
		opName = "CreateFile"
	}

	return &Error{code: code, err: &os.PathError{
		Op:   opName,
		Path: p,
		Err:  syscall.ENOENT,
	}}
}
