package filecache

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Cache struct {
	buf     map[string]*strings.Builder
	readAny bool
	prefix  string
	length  int
	mu      *sync.RWMutex
}

func ReadDir(dir, pattern string) (*Cache, error) {
	return ReadDirContext(context.Background(), dir, pattern)
}

func ReadDirContext(ctx context.Context, dir, pattern string) (*Cache, error) {
	c := &Cache{
		buf:    make(map[string]*strings.Builder),
		prefix: dir,
		mu:     &sync.RWMutex{},
	}
	if err := c.readDir(ctx, dir, pattern); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Cache) Get(name string) string {
	c.mu.RLock()
	s, ok := c.buf[filepath.Join(c.prefix, name)]
	c.mu.RUnlock()
	if !ok {
		return ""
	}
	return s.String()
}

func (c *Cache) Len() int {
	return c.length
}

func (c *Cache) Range(fn func(k, v string)) {
	for k, v := range c.buf {
		fn(k, v.String())
	}
}

func (c *Cache) check(dir, pattern string, ff os.FileInfo) error {
	name := ff.Name()
	fullName := filepath.Join(dir, name)
	if ff.IsDir() {
		return <-c.walk(fullName, pattern)
	}
	f, err := os.Open(fullName)
	if err != nil {
		return err
	}

	ok, err := filepath.Match(pattern, name)
	if err != nil {
		return err
	}
	if ok {
		return c.copy(f)
	}
	return nil
}

func (c *Cache) copy(f *os.File) error {
	var bd strings.Builder
	if _, err := io.Copy(&bd, f); err != nil {
		return err
	}
	c.mu.Lock()
	c.buf[f.Name()] = &bd
	c.mu.Unlock()
	c.length++
	return nil
}

func (c *Cache) readDir(ctx context.Context, dir, pattern string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c.walk(dir, pattern):
		return err
	}
}

func (c *Cache) walk(dir, pattern string) <-chan error {
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
			if err := c.check(dir, pattern, ff); err != nil {
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
