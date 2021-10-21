package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Galzzly/extract"
)

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
	destDir  string
	numC     int
)

func init() {
	flag.Var(&fileList, "f", "")
	flag.StringVar(&destDir, "d", "./", "")
	flag.IntVar(&numC, "c", 4, "")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() (err error) {
	if len(os.Args) >= 2 &&
		(os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help") {
		fmt.Println(usageString())
	}

	flag.Parse()

	/*
		Check that the destination directory exists
		If it doesn't, attempt to create it.
		Default is the current directory: ./
	*/
	if _, e := os.Stat(destDir); e != nil {
		if e := os.MkdirAll(destDir, 0755); e != nil {
			return e
		}
	}

	start := time.Now()
	//var wg sync.WaitGroup

	/*
		Check whether there are entries in fileList
		if there aren't, then populate with files
		from the current directory.
	*/
	if len(fileList) == 0 {
		fileList, err = getFileList()
		if err != nil {
			return
		}
	}

	err = extract.Extract(fileList, destDir, numC)
	if err != nil {
		return err
	}

	fmt.Println("Total Time:", time.Since(start))
	fmt.Println("Bundles have been extracted to:", destDir)
	return nil
}

func getFileList() (fileList []string, err error) {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		return
	}
	for _, f := range files {
		fileList = append(fileList, f.Name())
	}
	return
}

func usageString() string {
	buf := new(bytes.Buffer)
	buf.WriteString(usage)
	flag.CommandLine.SetOutput(buf)
	return buf.String()
}

const usage = `Usage: extract {help} {-f [filename]}
	help
		Display this help text. (Also -h or --help)
	
	-f [filename]
		To decompress a single bundle. May be used 
		more than once for multiple bundles.

	-d [directory]
		OPTIONAL. A target directory to extract the 
		bundles into. 

	-c [integer]
		OPTIONAL. Number of bundles to extract in 
		parallel. (Default: 4)
	
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
