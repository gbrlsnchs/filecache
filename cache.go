package filecache

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/gbrlsnchs/radix"
)

// Cache is an in-memory file cache.
type Cache struct {
	mu      *sync.RWMutex
	sem     chan struct{}
	tr      *radix.Tree
	readAny bool
	dir     string
	size    int64
	length  int
}

// New returns a file cache ready to read directories.
// The semaphore size is the number of CPUs.
func New(dir string) *Cache {
	return NewSize(dir, runtime.NumCPU())
}

// NewSize returns a file cache with a custom
// semaphore size ready to read directories.
func NewSize(dir string, size int) *Cache {
	return &Cache{
		mu:  &sync.RWMutex{},
		sem: make(chan struct{}, size),
		tr:  radix.New(radix.Tsafe),
		dir: filepath.Join(dir),
	}
}

// ReadDir reads a directory recursively and creates a cache
// of all files that match the expr. Errors if dir or any files under it can't be read.
//
// Files are cached in a hash map and its buffered content can be read using
// their path without the parent directory included as the key.
func ReadDir(dir, expr string) (*Cache, error) {
	return ReadDirContext(context.Background(), dir, expr)
}

// ReadDirContext is the context-aware equivalent of ReadDir.
// If a context gets done before it finishes caching all files, it returns an error.
func ReadDirContext(ctx context.Context, dir, expr string) (*Cache, error) {
	c := New(dir)
	if err := c.LoadContext(ctx, expr); err != nil {
		return nil, err
	}
	return c, nil
}

// Get returns a buffered file content.
func (c *Cache) Get(name string) string {
	n, _ := c.tr.Get(filepath.Join(c.dir, name))
	if n != nil && n.Value != nil {
		return n.Value.(*strings.Builder).String()
	}
	return ""
}

// Len returns the number of files cached.
func (c *Cache) Len() int {
	defer c.mu.RUnlock()
	c.mu.RLock()
	return c.length
}

// Load traverses the set directory recursively to cache files that match a given regexp.
func (c *Cache) Load(expr string) error {
	return c.LoadContext(context.Background(), expr)
}

// LoadContext does the same as ReadDir but is context-aware.
func (c *Cache) LoadContext(ctx context.Context, expr string) error {
	r, err := regexp.Compile(expr)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return c.walk(ctx, c.dir, r)
	}
}

// Size returns the total size in bytes of all cached files.
func (c *Cache) Size() int {
	defer c.mu.RUnlock()
	c.mu.RLock()
	return int(c.size)
}

func (c *Cache) String() string {
	s := "files"
	length := c.Len()
	if length == 1 {
		s = s[:len(s)-1]
	}
	ss := "bytes"
	size := c.Size()
	if size == 1 {
		ss = ss[:len(ss)-1]
	}
	return fmt.Sprintf("\n%d %s, %d %s:%v", length, s, size, ss, c.tr)
}

func (c *Cache) check(ctx context.Context, dir string, r *regexp.Regexp, ff os.FileInfo) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		name := ff.Name()
		fullName := filepath.Join(dir, name)
		if ff.IsDir() {
			return c.walk(ctx, fullName, r)
		}
		f, err := os.Open(fullName)
		if err != nil {
			return err
		}

		if r.MatchString(name) {
			return c.copy(f)
		}
		return nil
	}
}

func (c *Cache) copy(f *os.File) error {
	var bd strings.Builder
	n, err := io.Copy(&bd, f)
	if err != nil {
		return err
	}
	c.tr.Add(f.Name(), &bd) // this is thread-safe

	c.mu.Lock()
	c.size += n
	c.length++
	c.mu.Unlock()
	return nil
}

func (c *Cache) walk(ctx context.Context, dir string, r *regexp.Regexp) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	e := make(chan error)
	var wg sync.WaitGroup
	for _, ff := range files {
		select {
		case c.sem <- struct{}{}:
			wg.Add(1)
			go func(ff os.FileInfo) {
				defer func() {
					<-c.sem
					wg.Done()
				}()
				if err := c.check(ctx, dir, r, ff); err != nil {
					select {
					case e <- err:
						cancel()
					default:
					}
				}
			}(ff)
		default:
			if err = c.check(ctx, dir, r, ff); err != nil {
				return err
			}
		}
	}

	wg.Wait()
	close(e)
	return <-e // will emit the first error caught or nil
}
