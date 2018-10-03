# filecache (In-memory file caching using Go)
[![Build status](https://travis-ci.org/gbrlsnchs/filecache.svg?branch=master)](https://travis-ci.org/gbrlsnchs/filecache)
[![Build status](https://ci.appveyor.com/api/projects/status/qewple44o5rffms9/branch/master?svg=true)](https://ci.appveyor.com/project/gbrlsnchs/filecache/branch/master)
[![Sourcegraph](https://sourcegraph.com/github.com/gbrlsnchs/filecache/-/badge.svg)](https://sourcegraph.com/github.com/gbrlsnchs/filecache?badge)
[![GoDoc](https://godoc.org/github.com/gbrlsnchs/filecache?status.svg)](https://godoc.org/github.com/gbrlsnchs/filecache)
[![Minimal version](https://img.shields.io/badge/minimal%20version-go1.10%2B-5272b4.svg)](https://golang.org/doc/go1.10)

## About
This package recursively walks through a directory and caches files that match a regexp into a [radix tree](https://en.wikipedia.org/wiki/Radix_tree).

Since it spawns one goroutine for each file / directory lookup, it is also context-aware, enabling all the process to return earlier when the context is done.

## Usage
Full documentation [here](https://godoc.org/github.com/gbrlsnchs/filecache).

### Installing
#### Go 1.10
`vgo get -u github.com/gbrlsnchs/filecache`
#### Go 1.11 or after
`go get -u github.com/gbrlsnchs/filecache`

### Importing
```go
import (
	// ...

	"github.com/gbrlsnchs/filecache"
)
```

### Reading all files in a directory
```go
c, err := filecache.ReadDir("foobar", "")
if err != nil {
	// If err != nil, directory "foobar" doesn't exist, or maybe one of the files
	// inside this directory has been deleted during the reading.
}
txt := c.Get("bazqux.txt")
log.Print(txt)
```

### Reading specific files in a directory
```go
c, err := filecache.ReadDir("foobar", `\.sql$`)
if err != nil {
	// ...
}
q := c.Get("bazqux.sql")
log.Print(q)
log.Print(c.Len())  // amount of files cached
log.Print(c.Size()) // total size in bytes
```

### Lazy-reading a directory
```go
c := filecache.New("foobar")

// do stuff...

if err := c.Load(`\.log`); err != nil {
	// ...
}
```

### Setting a custom goroutine limit
By default, this package spawns goroutines for each file inside each directory.  
Currently, the limit of goroutines is the result of `runtime.NumCPU()`. However, it is possible to use a cache with a custom limit by using `filecache.NewSize` instead of `filecache.New`.
```go
c := filecache.NewSize("foobar", 100)

// do stuff...

if err := c.Load(`\.log`); err != nil {
	// ...
}
```

## Contributing
### How to help
- For bugs and opinions, please [open an issue](https://github.com/gbrlsnchs/filecache/issues/new)
- For pushing changes, please [open a pull request](https://github.com/gbrlsnchs/filecache/compare)
