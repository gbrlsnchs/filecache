package filecache_test

import (
	"context"
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
		cancel     bool
	}{
		{isNotExist: true},
		{dir: "testdata", length: 1},
		{dir: "testdata", err: context.Canceled, cancel: true},
	}
	for _, tc := range testCases {
		t.Run(tc.dir, func(t *testing.T) {
			ctx := context.Background()
			if tc.cancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			buf, err := ReadDirContext(ctx, tc.dir)
			if e, ok := err.(*os.PathError); ok {
				if want, got := tc.isNotExist, os.IsNotExist(e); want != got {
					t.Errorf("want %t, got %t", want, got)
				}
			} else if want, got := tc.err, err; want != got {
				t.Errorf("want %v, got %v", want, got)
			}
			if buf != nil {
				if want, got := tc.length, buf.Len(); want != got {
					t.Errorf("want %d, got %d", want, got)
				}
				buf.Range(func(k, v string) {
					t.Log(k)
					t.Log(v)
				})
			}
		})
	}
}
