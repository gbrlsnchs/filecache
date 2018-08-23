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
	exts    []string
	readAny bool
	prefix  string
	length  int
	mu      *sync.RWMutex
}

func ReadDir(dir string, exts ...string) (*Cache, error) {
	return ReadDirContext(context.Background(), dir, exts...)
}

func ReadDirContext(ctx context.Context, dir string, exts ...string) (*Cache, error) {
	c := &Cache{
		buf:     make(map[string]*strings.Builder),
		exts:    exts,
		readAny: len(exts) == 0,
		prefix:  dir,
		mu:      &sync.RWMutex{},
	}
	if err := c.readDir(ctx, dir); err != nil {
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

func (c *Cache) check(dir string, ff os.FileInfo) error {
	name := ff.Name()
	fullName := filepath.Join(dir, name)
	if ff.IsDir() {
		return <-c.walk(fullName)
	}
	f, err := os.Open(fullName)
	if err != nil {
		return err
	}

	if c.readAny {
		return c.copy(f)
	}
	ext := filepath.Ext(name)
	for i := range c.exts {
		if c.exts[i] == ext {
			return c.copy(f)
		}
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

func (c *Cache) readDir(ctx context.Context, dir string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c.walk(dir):
		return err
	}
}

func (c *Cache) walk(dir string) <-chan error {
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
			if err := c.check(dir, ff); err != nil {
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
