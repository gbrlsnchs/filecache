package filecache_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	. "github.com/gbrlsnchs/filecache"
)

var allFiles = []string{
	"lorem.txt", "main.go", "test.go", "test.log",
	"dir1/lorem.txt", "dir1/main.go", "dir1/test.go", "dir1/test.log",
	"dir2/lorem.txt", "dir2/main.go", "dir2/test.go", "dir2/test.log",
	"dir3/lorem.txt", "dir3/main.go", "dir3/test.go", "dir3/test.log",
	"dir4/lorem.txt", "dir4/main.go", "dir4/test.go", "dir4/test.log",
	"dir5/lorem.txt", "dir5/main.go", "dir5/test.go", "dir5/test.log",
}

func TestCache(t *testing.T) {
	testCases := []struct {
		dir        string
		expr       string
		cancel     bool
		files      []string
		err        error
		isNotExist bool
	}{
		{expr: ".", isNotExist: true},
		{dir: "testdata", expr: ".", files: allFiles},
		{dir: "testdata", expr: "", files: allFiles},
		{dir: "testdata", expr: `\.txt$`, files: filter(allFiles, `\.txt$`)},
		{dir: "testdata", expr: `\.go$`, files: filter(allFiles, `.go$`)},
		{dir: "testdata", expr: `\.(go|log)$`, files: filter(allFiles, `\.(go|log)$`)},
		{dir: "testdata", expr: `main\.+`, files: filter(allFiles, `main\.+`)},
		{dir: "testdata", expr: `test\.+`, files: filter(allFiles, `test\.+`)},
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

func filter(s []string, expr string) (f []string) {
	r, _ := regexp.Compile(expr)
	for _, ss := range s {
		if r.MatchString(ss) {
			f = append(f, ss)
		}
	}
	return f
}
