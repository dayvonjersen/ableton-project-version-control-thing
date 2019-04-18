/*
TODO(tso): create UI concept I've outlined on a piece of paper
*/

package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	gitIgnoreFile = `*
!*.xml
`
	gitPostCheckoutHook = `#!/bin/bash
for f in *.xml; do cat "$f" | gzip > "${f%%.xml}.als"; done
`
)

func gunzip(src, dest string) error {
	inFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inFile.Close()

	gz, err := gzip.NewReader(inFile)
	if err != nil {
		return err
	}
	defer gz.Close()

	outFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, gz)
	return err
}

func _gzip(src, dest string) error {
	outFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gz := gzip.NewWriter(outFile)
	defer gz.Close()

	inFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inFile.Close()

	_, err = io.Copy(gz, inFile)
	return err
}

func gitInit(filename string) {
	dir, _ := splitFilename(filename)
	gitDir := dir + "/.git"

	if fileExists(gitDir) {
		return
	}

	shellExec(dir, "git", "init")
	filePutContents(gitDir+"/info/exclude", gitIgnoreFile)
	filePutContents(gitDir+"/hooks/post-checkout", gitPostCheckoutHook)
}

func gitOnMasterBranch(filename string) bool {
	dir, _ := splitFilename(filename)

	head := fileGetContents(dir + "/.git/HEAD")
	if strings.HasPrefix(head, "ref: refs/heads/") {
		return strings.TrimSpace(strings.TrimPrefix(head, "ref: refs/heads/")) == "master"
	}

	return false
}

func gitCommit(filename string) {
	gitInit(filename)
	if !gitOnMasterBranch(filename) {
		return
	}

	xmlFilename := strings.TrimSuffix(normalizePathSeparators(filename), ".als") + ".xml"
	checkErr(gunzip(filename, xmlFilename))

	dir, xmlFilename := splitFilename(xmlFilename)

	shellExec(dir, "git", "add", xmlFilename)
	shellExec(dir, "git", "commit", "-m", "", "--allow-empty-message")
}

func gitAmend(filename, msg string) {
	if !gitOnMasterBranch(filename) {
		return
	}

	dir, _ := splitFilename(filename)
	shellExec(dir, "git", "commit", "--amend", "-m", msg, "--allow-empty-message")
}

func main() {
	// TODO(tso): flags?

	// NOTE(tso): this commits the latest version of all existing .als files
	//            creating new repos where necessary
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !strings.Contains(path, ".git") {
			path = normalizePathSeparators(path)
			if !dirExists(path + "/.git") {
				dir, err := os.Open(path)
				checkErr(err)
				files, err := dir.Readdir(-1)
				checkErr(err)
				for _, f := range files {
					if filepath.Ext(f.Name()) == ".als" {
						gitCommit(path + "/" + f.Name())
					}
				}
				dir.Close()
			}
		}
		return err
	})

	// NOTE(tso): this watches the current directory and all subdirectories for changes
	//            and automatically commits all changes to .als files
	//            every time the file is written (saved)
	var (
		lastFilename   string
		lastFilenameMu sync.Mutex
		// NOTE(tso): this mutex is necessary because the fsnotify stuff runs async
		// and i just don't want to run into problems.
		//
		// if the dispatcher calls the callback function synchronously instead,
		// it will throw off the timing with the fsnotify events, because
		// the callback function will get called for every event which actually
		// occurred in a < 100ms window because git is very slow
		//
		// -tso 2019-04-18 01:43:35a
	)

	w, err := newWatcher(
		func(filename string) bool {
			// fmt.Println("validator got:", filename)
			return filepath.Ext(filename) == ".als"
		},
		func(filename string) {
			fmt.Println(" callback got:", filename)
			gitCommit(filename)
			lastFilenameMu.Lock()
			defer lastFilenameMu.Unlock()
			lastFilename = filename
		},
	)
	checkErr(err)

	w.AddWithSubdirs(".")

	// NOTE(tso):  this lets you change the last commit message
	// by typing into the console while this program is running
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lastFilenameMu.Lock()
		if lastFilename != "" {
			gitAmend(lastFilename, scanner.Text())
		}
		lastFilenameMu.Unlock()
	}
}
