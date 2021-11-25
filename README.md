# extract ![issues](https://img.shields.io/github/issues/Galzzly/extract?style=plastic) [![extract GoDoc](https://img.shields.io/badge/reference-godoc-blue.svg?logo=go&style=plastic)](https://pkg.go.dev/github.com/Galzzly/extract/v2)

Introducing **extract v2.0** - a simple utility and GO library to extract different archive types. 

___
## Features

Package extract attempts to make it simple to decompress the compatible archive formats.

The `extract` command can be ran with no flags to decompress all compatible archive bundles in the current directory in place. It can also be ran with the following flags:


>`-f FILE | --flag=FILE` <br>to decomress a single bundle. This flag may be used more than once if there are multiple bundles.
><br>
>`-d DIR | --dest=DIR` <br>Sets a destination directory for the archives to be extracted into. By default this is set to the current working directory.
><br>
>`-c INT | --count=INT` <br>Sets the number of concurrent extractions that can take place. By default this is set to 4.
><br>
>`-h | --help` <br>Displays the help text
><br>
>`--version` <br>Displays the version of extract in use.

The `extract` tool now has the ability to run with concurrent extractions, by default 4. Previous versions of the tool ran in serial which was somewhat slow.

### Supported formats

The following archive/compression types are supported by extract:
- gzip
- tar
- zip
- rar (without password)
- bzip
- More to be added...

---
## GoDoc

See <https://pkg.go.dev/github.com/Galzzly/extract>

---
## Install

### GO

To install the binary directly into your \$GOPATH/bin:

```
go get github.com/Galzzly/extract/cmd/extract/v2
```
