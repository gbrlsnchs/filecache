package filecache_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/gbrlsnchs/filecache"
)

func TestBuffer(t *testing.T) {
	testCases := []struct {
		dir        string
		expr       string
		cancel     bool
		files      []string
		err        error
		isNotExist bool
	}{
		{expr: ".", isNotExist: true},
		{dir: "testdata", expr: ".", files: []string{"lorem.txt", "main.go", "test.go", "test.log"}},
		{dir: "testdata", expr: "", files: []string{"lorem.txt", "main.go", "test.go", "test.log"}},
		{dir: "testdata", expr: `\.txt$`, files: []string{"lorem.txt"}},
		{dir: "testdata", expr: `\.go$`, files: []string{"main.go", "test.go"}},
		{dir: "testdata", expr: `\.(go|log)$`, files: []string{"main.go", "test.go", "test.log"}},
		{dir: "testdata", expr: `^main\.+`, files: []string{"main.go"}},
		{dir: "testdata", expr: `^test\.+`, files: []string{"test.log", "test.go"}},
		{dir: "testdir", isNotExist: true},
		{dir: "testdata", expr: ".", err: context.Canceled, cancel: true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s/%s", tc.dir, tc.expr), func(t *testing.T) {
			ctx := context.Background()
			if tc.cancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			c, err := ReadDirContext(ctx, tc.dir, tc.expr)
			if e, ok := err.(*os.PathError); ok {
				if want, got := tc.isNotExist, os.IsNotExist(e); want != got {
					t.Errorf("want %t, got %t", want, got)
				}
			} else if want, got := tc.err, err; want != got {
				t.Errorf("want %v, got %v", want, got)
			}
			if c != nil {
				t.Log(c.String())

				if want, got := len(tc.files), c.Len(); want != got {
					t.Errorf("want %d, got %d", want, got)
				}
				var size int
				for _, fname := range tc.files {
					b, err := ioutil.ReadFile(filepath.Join(tc.dir, fname))
					if want, got := (error)(nil), err; want != got {
						t.Errorf("want %v, got %v", want, got)
					}
					size += len(b)
					if want, got := string(b), c.Get(fname); want != got {
						t.Errorf("want %s, got %s", want, got)
					}
				}
				if want, got := size, c.Size(); want != got {
					t.Errorf("want %d, got %d", want, got)
				}
			}
		})
	}
}
