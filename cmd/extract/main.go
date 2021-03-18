package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Galzzly/extract"
	"github.com/h2non/filetype"
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
)

func init() {
	flag.Var(&fileList, "f", "")
	flag.StringVar(&destDir, "d", "./", "")
}

func main() {
	start := time.Now()
	if len(os.Args) >= 2 &&
		(os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help") {
		fmt.Println(usageString())
	}

	if len(os.Args) == 1 {
		files, e := ioutil.ReadDir("./")
		extract.Check(e)
		for _, f := range files {
			fileList = append(fileList, f.Name())
		}
	} else {
		flag.Parse()
	}

	/*
		Check that the destination directory exists
		If it doesn't, attempt to create it.
		Default is the current directory: ./
	*/
	if _, e := os.Stat(destDir); e != nil {
		if e := os.MkdirAll(destDir, 0755); e != nil {
			panic(e)
		}
	}

	for _, v := range fileList {
		// Check the file, and make sure we know what it is.
		buf, _ := ioutil.ReadFile(v)
		format, _ := filetype.Match(buf)

		// If we do, send it on for processing
		if format != filetype.Unknown {
			extr(v, format.Extension, format.MIME.Value)
		}

	}

	fmt.Println("Total Time:", time.Since(start))
	fmt.Println("Bundles have been extracted to:", destDir)
}

func extr(source, ext, frmt string) {
	var err error
	fmt.Printf("Looking at archive %s ... ", source)
	strtExtr := time.Now()
	switch frmt {
	// Gzip files
	case "application/gzip":
		//fmt.Println("gzip")
		f, e := os.Open(source)
		extract.Check(e)
		defer f.Close()
		err = extract.Gzip(f, destDir)
	// Tar files
	case "application/x-tar":
		//fmt.Println("tar")
		f, e := os.Open(source)
		extract.Check(e)
		defer f.Close()
		err = extract.Tar(f, destDir)
	// Zip files
	case "application/zip":
		//fmt.Println("zip")
		err = extract.Zip(source, destDir)
	// 7-Zip files
	//case "application/x-7z-compressed":
	//fmt.Println("7za")
	// Rar files, without password
	case "application/x-rar", "application/vnd.rar":
		f, e := os.Open(source)
		extract.Check(e)
		defer f.Close()
		err = extract.Rar(f, destDir)
	// Bzip files
	case "application/x-bzip2":
		//fmt.Println("bzip")
		f, e := os.Open(source)
		extract.Check(e)
		defer f.Close()
		err = extract.Bzip(f, extract.GetFileName(source), destDir)
	// Anything else, we do not process right now
	default:
		fmt.Println("Unable to process right now")
	}

	if err != nil {
		fmt.Println("Failed in", time.Since(strtExtr))
	} else {
		fmt.Println("Successful in", time.Since(strtExtr))
	}
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
