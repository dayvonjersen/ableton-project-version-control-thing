/*
TODO(tso):
    - update readme
    - create UI concept I've outlined on a piece of paper
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

func gitLog(filename string) {
	dir, _ := splitFilename(filename)
	shellExecAttach(dir, "git", "log", "--stat", `--format=%h %cr %s`)
}

func gitCheckout(filename, revision string) {
	dir, fname := splitFilename(filename)
	xmlFilename := strings.TrimSuffix(normalizePathSeparators(fname), ".als") + ".xml"
	shellExec(dir, "git", "checkout", "-b", "temp")
	shellExec(dir, "git", "checkout", revision, "--", xmlFilename)
}

func gitMerge(filename string) {
	dir, _ := splitFilename(filename)
	// NOTE(tso): the post-checkout hook takes care of gzipping the xml file for us
	// NOTE(tso): the post-checkout hook takes care of gzipping the xml file for us
	// NOTE(tso): the post-checkout hook takes care of gzipping the xml file for us
	//xmlFilename := strings.TrimSuffix(normalizePathSeparators(filename), ".als") + ".xml"
	//checkErr(gunzip(filename, xmlFilename))

	// NOTE(tso): "git checkout revision -- file" automatically adds the file to the index for us
	// NOTE(tso): "git checkout revision -- file" automatically adds the file to the index for us
	// NOTE(tso): "git checkout revision -- file" automatically adds the file to the index for us
	//shellExec(dir, "git", "add", xmlFilename)
	shellExec(dir, "git", "commit", "-m", "[AUTO] revert "+filename, "--allow-empty-message")
	shellExec(dir, "git", "checkout", "master")
	shellExec(dir, "git", "merge", "temp")
	shellExec(dir, "git", "branch", "-D", "temp")
}

func gitReset(filename string) {
	dir, fname := splitFilename(filename)
	xmlFilename := strings.TrimSuffix(normalizePathSeparators(fname), ".als") + ".xml"
	shellExec(dir, "git", "reset", "HEAD")
	shellExec(dir, "git", "checkout", xmlFilename)
	shellExec(dir, "git", "checkout", "master")
	shellExec(dir, "git", "branch", "-D", "temp")
}

func main() {
	fmt.Println("a b l e t o n  +  g i t = <3")
	fmt.Println("-------------------------")
	fmt.Println()
	fmt.Println("This software is EXPERIMENTAL.")
	fmt.Println()
	fmt.Println("MAKE BACKUPS OF YOUR PROJECTS BEFORE USING THIS SOFTWARE")
	fmt.Println("MAKE BACKUPS OF YOUR PROJECTS WHILE USING THIS SOFTWARE")
	fmt.Println("MAKE BACKUPS OF YOUR PROJECTS AFTER USING THIS SOFTWARE")
	fmt.Println()
	fmt.Println("The author takes no responsibility for any loss of data incurred while using this software.")
	fmt.Println()
	fmt.Println("Report any issues you may encounter, or any ideas you might have for improving this thing to")
	fmt.Println("https://github.com/dayvonjersen/ableton-project-version-control-thing/issues")
	fmt.Println()
	fmt.Println("-------------------------")

	// TODO(tso): flags?

	cwd, err := os.Getwd()
	checkErr(err)
	_, cwd = splitFilename(cwd)

	// NOTE(tso): this commits the latest version of all existing .als files
	//            creating new repos where necessary
	fmt.Println("Scanning", cwd, "and all subdirs for existing files:")
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
						filename := path + "/" + f.Name()
						fmt.Println("Found existing project:", filename)
						gitInit(filename)
						if gitOnMasterBranch(filename) {
							gitCommit(filename)
						}
					}
				}
				dir.Close()
			} else {
				fmt.Println("Found existing repo at:", path)
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

		doingGitStuffMu sync.Mutex
		doingGitStuff   bool
		// NOTE(tso): don't want to be auto-committing before/during/after a merge
		// -tso 2019-04-23 02:05:45a
	)

	fmt.Println("Watching", cwd, "for changes to files...")
	w, err := newWatcher(
		func(filename string) bool {
			// fmt.Println("validator got:", filename)
			return filepath.Ext(filename) == ".als"
		},
		func(filename string) {
			fmt.Println("Got file change:", filename)
			doingGitStuffMu.Lock()
			defer doingGitStuffMu.Unlock()
			if doingGitStuff {
				return
			}

			gitInit(filename)
			if gitOnMasterBranch(filename) {
				gitCommit(filename)
			}
			lastFilenameMu.Lock()
			defer lastFilenameMu.Unlock()
			lastFilename = filename
		},
	)
	checkErr(err)

	w.AddWithSubdirs(".")

	fmt.Println()
	fmt.Println("Ready.")
	help()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()

		lastFilenameMu.Lock()
		last := lastFilename
		lastFilenameMu.Unlock()

		args := strings.SplitN(text, " ", 2)
		switch strings.ToLower(args[0]) {
		case "": // do nothing
		case "help", "h", "?":
			help()

		case "current":
			lastFilenameMu.Lock()
			if lastFilename != "" {
				fmt.Println("[INFO] current file is:", lastFilename)
			} else {
				fmt.Println("[INFO] current file is not set.")
			}
			lastFilenameMu.Unlock()
		case "set":
			if len(args) == 2 {
				if fileExists(args[1]) {
					lastFilenameMu.Lock()
					lastFilename = args[1]
					lastFilenameMu.Unlock()
					fmt.Println("[ OK ] current file set to:", args[1])
				}
			} else {
				fmt.Println("[INFO] usage: set [FILENAME]")
			}
		case "log":
			if checkLast(last) {
				gitLog(last)
			}
		case "checkout":
			if len(args) == 2 && checkLast(last) {
				doingGitStuffMu.Lock()
				doingGitStuff = true
				doingGitStuffMu.Unlock()

				gitCheckout(last, args[1])
			} else {
				fmt.Println("[INFO] usage: checkout [HASH]")
			}
		case "save":
			if checkLast(last) {
				gitMerge(last)

				doingGitStuffMu.Lock()
				doingGitStuff = false
				doingGitStuffMu.Unlock()
			}
		case "cancel":
			if checkLast(last) {
				gitReset(last)

				doingGitStuffMu.Lock()
				doingGitStuff = false
				doingGitStuffMu.Unlock()
			}
		default:
			// NOTE(tso):  this lets you change the last commit message
			// by typing into the console while this program is running
			if checkLast(last) {
				gitAmend(last, text)
			}
		}
	}
}

func checkLast(last string) bool {
	if last == "" {
		fmt.Println("[OHNO] current file is not set")
		return false
	}
	return true
}

func help() {
	fmt.Println()
	fmt.Println("The following commands are available:")
	fmt.Println("current")
	fmt.Println("set [FILENAME]")
	fmt.Println("log")
	fmt.Println("checkout [HASH]")
	fmt.Println("save")
	fmt.Println("cancel")
	fmt.Println()
	fmt.Println("to change the most recent commit message, simply type it and press enter")
}
