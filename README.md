# filecache (In-memory file caching using Go)
[![Build Status](https://travis-ci.org/gbrlsnchs/filecache.svg?branch=master)](https://travis-ci.org/gbrlsnchs/filecache)
[![GoDoc](https://godoc.org/github.com/gbrlsnchs/filecache?status.svg)](https://godoc.org/github.com/gbrlsnchs/filecache)

## About
This package recursively walks through a directory and caches files that match a regexp into a map of `strings.Builder` buffers. Those string builders can be accessed through the file path name under the main directory.

Since it spawns one goroutine for each file/directory, it is also context-aware, enabling all the process to return earlier when the context is done.

## Usage
Full documentation [here].

## Example
### Reading all files in a directory
```go
c, err := filecache.ReadDir("foobar", ".")
if err != nil {
	// If err != nil, directory "foobar" doesn't exist, or maybe one of the files
	// inside this directory has been deleted just after it has been listed.
}
txt := c.Get("bazqux")
log.Print(txt)
```

### Reading specific files in a directory
```go
c, err := filecache.ReadDir("foobar", `\.sql$`)
if err != nil {
	// ...
}
q := c.Get("query.sql")
log.Print(q)
log.Print(c.Count()) // amount of files cached
log.Print(c.Len())   // total size in bytes
```

## Contribution
### How to help:
- Pull Requests
- Issues
- Opinions

[here]: https://godoc.org/github.com/gbrlsnchs/filecache