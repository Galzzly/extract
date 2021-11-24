package main

import (
	"fmt"
	"io/ioutil"
	"os"

	extract "github.com/Galzzly/extract/v2"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	fileList = kingpin.Flag("file", "To decompress a single bundle. May be used more than once for multiple bundles.").Short('f').Strings()
	destDir  = kingpin.Flag("dest", "Destination directory for the decompressed bundle.").Short('d').Default("./").String()
	numC     = kingpin.Flag("count", "Number of concurrent extractions.").Short('c').Default("4").Uint32()
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() (err error) {
	kingpin.Version("2.0.0")
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	/*
		If no files are specified, attempt to get a list of files in the current directory.
	*/

	if len(*fileList) == 0 {
		fileList, err = getFileList()
		if err != nil {
			return err
		}
	}

	/*
		Check that the destination directory exists
		If it doesn't, attempt to create it.
		Default is the current directory: ./
	*/
	if _, err := os.Stat(*destDir); err != nil {
		if err := os.MkdirAll(*destDir, 0755); err != nil {
			return err
		}
	}

	/*
		Extract the files
	*/
	if err := extract.Extract(fileList, *destDir, *numC); err != nil {
		return err
	}

	return nil
}

func getFileList() (fileList *[]string, err error) {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		return
	}
	var fl = make([]string, 0, len(files))
	for _, f := range files {
		fl = append(fl, f.Name())
	}
	fileList = &fl
	return
}
