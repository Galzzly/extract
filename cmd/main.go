package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/h2non/filetype"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	fileList arrayFlags
)

func init() {
	flag.Var(&fileList, "f", "")
}

func main() {
	if len(os.Args) >= 2 &&
		(os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help") {
		fmt.Println(usageString())
	}

	if len(os.Args) == 1 {
		files, e := ioutil.ReadDir("./")
		check(e)
		for _, f := range files {
			fileList = append(fileList, f.Name())
		}
	} else {
		flag.Parse()
	}

	for _, v := range fileList {
		// Check the file
		buf, _ := ioutil.ReadFile(v)
		format, _ := filetype.Match(buf)
		if format != filetype.Unknown {
			// fmt.Printf("File type: %s, MIME: %s\n", format.Extension, format.MIME.Value)
			extract(v, format.Extension, format.MIME.Value)
		}

	}
}

func extract(source, ext, frmt string) {
	fmt.Printf("Extracting archive: %s\n", source)
	switch frmt {
	case "application/gzip":
		fmt.Println("gzip")
	case "application/x-tar":
		fmt.Println("tar")
	case "application/zip":
		fmt.Println("zip")
	case "application/x-7z-compressed":
		fmt.Println("7za")
	case "application/x-rar":
		fmt.Println("rar")
	case "application/x-bzip2":
		fmt.Println("bzip")
	}
}

func usageString() string {
	buf := new(bytes.Buffer)
	buf.WriteString(usage)
	flag.CommandLine.SetOutput(buf)
	flag.CommandLine.PrintDefaults()
	return buf.String()
}

const usage = `Usage: extract {help} {-f [filename]}
	help
		Display this help text. (Also -h or --help)
	
	-f [filename]
		To decompress a single bundle. May be used 
		more than once for multiple bundles.

	Without any arguments, the utility will iterate 
	through all of the files in the current directory 
	and extract where possible.

	Archive Formats Supported
		.tar
		.tar.gz
		.gz
		.zip
		.rar
		.bz2
`
