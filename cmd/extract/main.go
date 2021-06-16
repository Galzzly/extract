package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
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
	var wg sync.WaitGroup

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
			panic(e)
		}
	}

	/*
		Check whether there are entries in fileList
		if there aren't, then populate with files
		from the current directory.
	*/
	if len(fileList) == 0 {
		files, e := ioutil.ReadDir("./")
		extract.Check(e)
		for _, f := range files {
			fileList = append(fileList, f.Name())
		}
	}

	sem := make(chan int, 4) // Up to 4 jobs at once
	wg.Add(len(fileList))

	for _, v := range fileList {
		go worker(v, &wg, sem)
	}
	wg.Wait()
	close(sem)

	fmt.Println("Total Time:", time.Since(start))
	fmt.Println("Bundles have been extracted to:", destDir)
}

func worker(source string, wg *sync.WaitGroup, sem chan int) {
	defer wg.Done()
	sem <- 1
	buf, _ := ioutil.ReadFile(source)
	format, _ := filetype.Match(buf)

	if filetype.IsArchive(buf) {
		fmt.Println("Looking at archive", source)
		strtExtr := time.Now()
		e := extr(source, format.Extension, format.MIME.Value)
		printTime(source, e, strtExtr)
	}

	<-sem
}

func extr(source, ext, frmt string) error {
	var err error
	//fmt.Printf("Looking at archive %s ... ", source)
	//strtExtr := time.Now()
	switch frmt {
	// Gzip files
	case "application/gzip":
		//fmt.Println("gzip")
		f, e := os.Open(source)
		extract.Check(e)
		defer f.Close()
		err = extract.Gzip(f, destDir)
		//printTime(err, strtExtr)
	// Tar files
	case "application/x-tar":
		//fmt.Println("tar")
		f, e := os.Open(source)
		extract.Check(e)
		defer f.Close()
		err = extract.Tar(f, destDir)
		//printTime(err, strtExtr)
	// Zip files
	case "application/zip":
		//fmt.Println("zip")
		err = extract.Zip(source, destDir)
		//printTime(err, strtExtr)
	// Rar files, without password
	case "application/x-rar", "application/vnd.rar":
		f, e := os.Open(source)
		extract.Check(e)
		defer f.Close()
		err = extract.Rar(f, destDir)
		//printTime(err, strtExtr)
	// Bzip files
	case "application/x-bzip2":
		//fmt.Println("bzip")
		f, e := os.Open(source)
		extract.Check(e)
		defer f.Close()
		err = extract.Bzip(f, extract.GetFileName(source), destDir)
		//printTime(err, strtExtr)
	// Anything else, we do not process right now
	default:
		// fmt.Println("Unable to process right now")
		err = fmt.Errorf("Unable to process right now")
	}
	/*
		if err != nil {
			fmt.Println("Failed in", time.Since(strtExtr))
		} else {
			fmt.Println("Successful in", time.Since(strtExtr))
		}
	*/

	return err
}

func printTime(source string, e error, t time.Time) {
	if e != nil {
		fmt.Println(source, "extract failed in", time.Since(t))
	} else {
		fmt.Println(source, "extract successful in", time.Since(t))
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
