package filecache_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	. "github.com/gbrlsnchs/filecache"
)

func TestBuffer(t *testing.T) {
	testCases := []struct {
		isNotExist bool
		length     int
		err        error
		dir        string
		pattern    string
		cancel     bool
	}{
		{isNotExist: true, pattern: "*"},
		{dir: "testdata", length: 2, pattern: "*"},
		{dir: "testdata", length: 1, pattern: "*.go"},
		{dir: "testdata", err: context.Canceled, cancel: true, pattern: "*"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s/**/%s", tc.dir, tc.pattern), func(t *testing.T) {
			ctx := context.Background()
			if tc.cancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			c, err := ReadDirContext(ctx, tc.dir, tc.pattern)
			if e, ok := err.(*os.PathError); ok {
				if want, got := tc.isNotExist, os.IsNotExist(e); want != got {
					t.Errorf("want %t, got %t", want, got)
				}
			} else if want, got := tc.err, err; want != got {
				t.Errorf("want %v, got %v", want, got)
			}
			if c != nil {
				if want, got := tc.length, c.Len(); want != got {
					t.Errorf("want %d, got %d", want, got)
				}
				c.Range(func(k, v string) {
					t.Log(k)
					t.Log(v)
				})
			}
		})
	}
}
