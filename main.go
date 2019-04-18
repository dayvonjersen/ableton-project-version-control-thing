/*
   TODO:
   1. fix the stuff i know isn't working correctly rn
       ✓ files with spaces in the names
       ✓ actually running when i hit ctrl+s in ableton

      get this working for a single directory before

   2. ensuring recursive behavior works correctly

      then

   3. create UI concept I've outlined on a piece of paper
*/

package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	shellExec(dir, "git", "commit", "-m", "", "--allow-empty-message", "--allow-empty")
	filePutContents(gitDir+"/info/exclude", gitIgnoreFile)
	filePutContents(gitDir+"/hooks/post-checkout", gitPostCheckoutHook)
}

func gitOnMasterBranch(filename string) bool {
	dir, _ := splitFilename(filename)

	head := shellExecString(dir, "git", "name-rev", "--name-only", "HEAD")
	return head == "master"
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
	w, err := newWatcher(
		func(filename string) bool {
			// fmt.Println("validator got:", filename)
			return filepath.Ext(filename) == ".als"
		},
		func(filename string) {
			fmt.Println(" callback got:", filename)
			gitCommit(filename)
		},
	)
	checkErr(err)

	w.AddWithSubdirs(".")
	select {}

	/*
		    this commits the latest version of all existing .als files
		    creating new repos where necessary

			filepath.Walk(watchDir, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() && !strings.Contains(path, ".git") {
					if !dirExists(path + "/.git") {
						dir, err := os.Open(path)
						checkErr(err)
						files, err := dir.Readdir(-1)
						checkErr(err)
						alsFiles := []string{}
						for _, f := range files {
							if filepath.Ext(f.Name()) == ".als" {
								alsFiles = append(alsFiles, f.Name())
							}
						}
						dir.Close()
						for _, alsFile := range alsFiles {
							f := &file{
								Name: alsFile,
								Dir:  path,
								Ext:  ".als",
								Time: time.Now().Unix(),
							}
							gitCommit(f)
						}
					}
					watchPath(path)
				}
				return err
			})
	*/

	/*
		    this keeps track of the last repo we worked in

			files := make(chan *file)
			rf := recentFiles{}
	*/

	/*
		    this lets you change the commit message by typing into the console when this program is running
			go func() {
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					if len(rf) > 0 {
						gitAmend(rf[0], scanner.Text())
					}
				}
			}()
	*/

	/*
		    this is the old dispatcher routine
			go func() {
				for {
					f := <-files
					i, ok := rf.Find(f)
					if !ok || rf[i].Time < f.Time {
						if ok {
							rf[i].Time = f.Time
						} else {
							rf = append(rf, f)
							sort.Sort(rf)
						}
						gitCommit(f)
					}
				}
			}()
			for {
				select {
				case e := <-w.Event:
					name := normalizePathSeparators(e.Name)
					dir := normalizePathSeparators(filepath.Dir(name))
					ext := filepath.Ext(name)
					if ext == ".als" {
						files <- &file{
							Name: name,
							Dir:  dir,
							Ext:  ext,
							Time: time.Now().Unix(),
						}
					} else if dirExists(name) && !strings.Contains(name, ".git") {
						watchPath(name)
					}
				case err := <-w.Error:
					checkErr(err)
				}
			}
	*/
}

/*
this probably isn't needed anymore

type recentFiles []*file

func (rf recentFiles) Len() int { return len(rf) }
func (rf recentFiles) Less(i, j int) bool {
	if rf[i].Name == rf[j].Name {
		return rf[i].Time > rf[j].Time
	}
	return rf[i].Name < rf[j].Name
}
func (rf recentFiles) Swap(i, j int) { rf[i], rf[j] = rf[j], rf[i] }
func (rf recentFiles) Find(f *file) (int, bool) {
	i := sort.Search(len(rf), func(i int) bool {
		return rf[i].Name == f.Name
	})
	return i, i < len(rf) && rf[i].Name == f.Name
}
*/
