package filecache

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// Cache is an in-memory file cache.
type Cache struct {
	buf     map[string]*strings.Builder
	readAny bool
	prefix  string
	size    int64
	mu      *sync.RWMutex
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
	c := &Cache{
		buf:    make(map[string]*strings.Builder),
		prefix: dir,
		mu:     &sync.RWMutex{},
	}
	r, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	if err := c.readDir(ctx, dir, r); err != nil {
		return nil, err
	}
	return c, nil
}

// Count returns the number of files cached.
func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.buf)
}

// Get returns a buffered file content.
func (c *Cache) Get(name string) string {
	c.mu.RLock()
	s, ok := c.buf[filepath.Join(c.prefix, name)]
	c.mu.RUnlock()
	if !ok {
		return ""
	}
	return s.String()
}

// Len returns the total size in bytes of all cached files.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return int(c.size)
}

// Range iterates over the map that holds the cached content.
func (c *Cache) Range(fn func(k, v string)) {
	for k, v := range c.buf {
		fn(k, v.String())
	}
}

func (c *Cache) check(dir string, r *regexp.Regexp, ff os.FileInfo) error {
	name := ff.Name()
	fullName := filepath.Join(dir, name)
	if ff.IsDir() {
		return <-c.walk(fullName, r)
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

func (c *Cache) copy(f *os.File) error {
	var bd strings.Builder
	n, err := io.Copy(&bd, f)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.buf[f.Name()] = &bd
	c.size += n
	c.mu.Unlock()
	return nil
}

func (c *Cache) readDir(ctx context.Context, dir string, r *regexp.Regexp) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c.walk(dir, r):
		return err
	}
}

func (c *Cache) walk(dir string, r *regexp.Regexp) <-chan error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		e := make(chan error, 1)
		e <- err
		close(e)
		return e
	}

	e := make(chan error)
	var wg sync.WaitGroup
	wg.Add(len(files))
	for _, ff := range files {
		go func(ff os.FileInfo) {
			defer wg.Done()
			if err := c.check(dir, r, ff); err != nil {
				select {
				case e <- err:
				default:
				}
			}
		}(ff)
	}

	go func() {
		wg.Wait()
		close(e)
	}()
	return e
}
