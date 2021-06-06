package main

import (
	"archive/zip"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	destination := flag.String("destination", "", "Specify destination directory and/or filename for the archive. If you don't specify the name of the destination, source directory's name will be used")
	source := flag.String("source", "", "Specify the path to the directory you want to archive")
	flag.Parse()

	// Check if source directory is specified
	if len(*source) == 0 {
		log.Fatalf("Missing required --source parameter")
	}

	// Check if source is a directory
	fileInfo, err := os.Stat(*source)
	if err != nil {
		log.Fatalln("There was an error checking the source directory\n" + err.Error())
	}
	if !fileInfo.IsDir() {
		log.Fatalf("%s is not a directory", *source)
	}

	// Get the source directory name
	s := strings.Split(*source, `/`) // Linux
	if len(s) == 1 {
		s = strings.Split(*source, `\`) // Windows
	}
	sourceDirectoryName := s[len(s)-2]

	// Generating a string of timestamp
	now := time.Now().Format("2006_01_02_15_04_00")

	// Checks if destination was specified, if not, source directory name will be used as base of destination file name
	if len(*destination) == 0 {
		*destination = sourceDirectoryName + "&" + now + ".zip"
	} else if strings.HasSuffix(*destination, ".zip") {
		*destination = strings.TrimSuffix(*destination, ".zip")
		*destination = *destination + "&" + now + ".zip"
	} else if !strings.HasSuffix(*destination, ".zip") {
		*destination = *destination + "/" + sourceDirectoryName + "&" + now + ".zip"
	}

	// Creating archive file
	file, err := os.Create(*destination)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	walker := func(path string, info os.FileInfo, err error) error {
		// Check if any error was passed to walker, and pass it back to main function
		if err != nil {
			return err
		}

		// Skip rest of the function if directory
		if info.IsDir() {
			return nil
		}

		// Opening source file
		file, err := os.Open(path)
		if err != nil {
			log.Printf("Unable to open file: %s\n", path)
			return err
		}
		defer file.Close()

		// Converting from path to zip-root relative path
		path = strings.TrimPrefix(path, *source)

		// Creating file in archive
		f, err := w.Create(path)
		if err != nil {
			log.Printf("There was an error creating file in archive: %s\n", path)
			return err
		}

		// Copying source file to zip io-writer
		_, err = io.Copy(f, file)
		if err != nil {
			log.Printf("There was an error copying file: %s\n", path)
			return err
		}

		return nil
	}

	// Walk through all files inside specified source directory including all sub directories
	err = filepath.Walk(*source, walker)
	if err != nil {
		panic(err)
	}
}
