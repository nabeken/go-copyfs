# go-copyfs

`go-copyfs` is a library to copy a given `fs.FS` into the local filesystem.

**Note**: `go-copyfs` doesn't support the symlink because `fs.FS` [doesn't support as of Go 1.18](https://github.com/golang/go/issues/49580).

# Motivation

When I was writing tests, I wanted to extract files in `embed.FS` into the local filesystem because the code only accept a filename in the local filesystem. I had to write some code to deal with the temporary directory and copy the files into it.

That's why I wrote this small library, `go-copyfs`.

# Usage

See [copyfs_test.go](copyfs_test.go)

# License

See [LICENSE](LICENSE)
