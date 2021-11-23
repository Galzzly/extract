# extract ![issues](https://img.shields.io/github/issues/Galzzly/extract?style=plastic) [![extract GoDoc](https://img.shields.io/badge/reference-godoc-blue.svg?logo=go&style=plastic)](https://pkg.go.dev/github.com/Galzzly/extract)

Introducing **extract v2.0** - a simple utility and GO library to extract different archive types. 

## Features

Package extract attempts to make it simple to decompress the compatible archive formats.

The `extract` command can be ran with no flags to decompress all compatible archive bundles in the current directory in place, or can be ran with `-f [archive-file]` to point to a specific archive bundle (multiple `-f` options can be used). A `-d [directory]` can be used to specify an output directoy. Optionally, a `-c [integer]` can be used to specify the number of extractions performed in parallel.

The `extract` utility will run up to four extracts concurrently. Previous versions of the tool ran in serial which was somewhat slow.

### Supported formats

The following archive/compression types are supported by extract:
- gzip
- tar
- zip
- rar (without password)
- bzip
- More to be added...

## GoDoc

See <https://pkg.go.dev/github.com/Galzzly/extract>

## Install

### GO

To install the binary directly into your \$GOPATH/bin:

```
go get github.com/Galzzly/extract/cmd/extract
```
