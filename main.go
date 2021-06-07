package main

import (
	"archive/zip"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	source        string
	destination   string
	w             *zip.Writer
	archiveNeeded = false
)

func splitPath(path string) []string {
	s := strings.Split(path, `/`) // Linux
	if len(s) == 1 {
		s = strings.Split(path, `\`) // Windows
	}
	return s
}

func isDirectory(path string) {
	// Check if source is a directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatalf("There was an error checking the directory '%s'! Make sure it exist\n", path)
	}
	if !fileInfo.IsDir() {
		log.Fatalf("%s is not a directory", path)
	}
}

// Sanitises inputs and checks source directory
func runChecks() {
	var (
		archiveNameBase string    // Used to check simmilarly names archives
		lastArchived    time.Time // Used to store the last archive's timestamp
	)
	// Check if source directory is specified
	if len(source) == 0 {
		log.Fatalf("Missing required --source parameter")
	}

	s := splitPath(source)
	sourceDirectoryName := s[len(s)-2]

	// Check and make sure source is a directory, script curently doesn't support single file to be archived
	isDirectory(source)

	// Generating a string of timestamp
	timeFormat := "2006_01_02T15_04"
	now := time.Now().Format(timeFormat)

	// If destination is not specified the name of the archive will be the same as the source's directory name, and the archive will be saved in the current directory
	// If .zip extention is used then that name will be used to generate the archive
	// If directory is specified as destination the name of the source directory will be used as the name of the archive, and will be saved to the specified destination directory
	if len(destination) == 0 {
		archiveNameBase = sourceDirectoryName
		destination = archiveNameBase + "&" + now + ".zip"
	} else if strings.HasSuffix(destination, ".zip") {
		destination = strings.TrimSuffix(destination, ".zip")
		s := splitPath(destination)
		archiveNameBase = s[len(s)-1]
		destination = destination + "&" + now + ".zip"
	} else if !strings.HasSuffix(destination, ".zip") {
		isDirectory(destination)
		if runtime.GOOS == "windows" {
			destination = destination + `\`
		} else {
			destination = destination + "/"
		}
		archiveNameBase = sourceDirectoryName
		destination = destination + sourceDirectoryName + "&" + now + ".zip"
	}

	// Get files in destination
	files, err := ioutil.ReadDir(filepath.Dir(destination))
	if err != nil {
		log.Fatal(err)
	}

	// Check each file in destnation directory. If the file name matches the new archive's base name the last modified timestamp of it will be stored.
	for _, fileInfo := range files {
		archiveFound, err := filepath.Match(archiveNameBase+"*", fileInfo.Name())
		if err != nil {
			log.Fatalf("There was an error matching archive %s\n", fileInfo.Name())
		}
		if archiveFound {
			// If newer archive is found it's timestamp will be recorded
			if lastArchived.Before(fileInfo.ModTime()) {
				lastArchived = fileInfo.ModTime()
			}
		}
	}

	sourceWalker := func(path string, info os.FileInfo, err error) error {
		// Check if any error was passed to walker, and pass it back to main function
		if err != nil {
			return err
		}

		// Skip rest of the function if directory
		if info.IsDir() {
			return nil
		}

		// Getting file info
		fileInfo, err := os.Stat(path)
		if err != nil {
			log.Fatalf("There was an error checking archive %s\n", path)
		}

		// If there is a file that was modified after the latest archive was created archive needed will be set to true
		if fileInfo.ModTime().After(lastArchived) {
			archiveNeeded = true
		}

		return nil
	}

	// Check if there are any older files in the source directory than the newest archive
	err = filepath.Walk(source, sourceWalker)
	if err != nil {
		panic(err)
	}
}

func archiveWalker(path string, info os.FileInfo, err error) error {
	// Check if any error was passed to walker, and pass it back to main function
	if err != nil {
		return err
	}

	// Skip rest of the function if directory
	if info.IsDir() {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		log.Printf("Unable to open file: %s\n", path)
		return err
	}
	defer file.Close()

	// Converting from path to zip-root relative path
	path = strings.TrimPrefix(path, source)

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

func main() {
	flag.StringVar(&destination, "destination", "", "Specify destination directory and/or filename for the archive. If you don't specify the name of the destination, source directory's name will be used")
	flag.StringVar(&source, "source", "", "Specify the path to the directory you want to archive")
	flag.Parse()

	runChecks()

	// Walk through all files inside of specified source directory including all sub directories
	if archiveNeeded {
		// Creating archive file
		file, err := os.Create(destination)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		w = zip.NewWriter(file)
		defer w.Close()
		// Archive each file in source directory
		err = filepath.Walk(source, archiveWalker)
		if err != nil {
			panic(err)
		}
		log.Printf("%s has been archived to %s", source, destination)
	} else {
		log.Println("There have been no changes in source since last archive")
	}
}
