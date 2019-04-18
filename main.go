/*
   TODO:
   1. fix the stuff i know isn't working correctly rn
       - files with spaces in the names
       *** actually running when i hit ctrl+s in ableton ***

      get this working for a single directory before

   2. ensuring recursive behavior works correctly

      then

   3. create UI concept I've outlined on a piece of paper
*/

package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type file struct {
	Name, Dir, Ext string
	Time           int64
}

func gitInit(f *file) {
	gitDir := f.Dir + "/.git"
	if fileExists(gitDir) {
		return
	}
	shellExec(f.Dir, "git", "init")
	shellExec(f.Dir, "git", "commit", "-m", "", "--allow-empty-message", "--allow-empty")
	filePutContents(gitDir+"/info/exclude", "*\n!*.xml\n")
	filePutContents(gitDir+"/hooks/post-checkout", "#!/bin/bash\nfor f in `ls --color=never *.xml`; do cat $f | gzip > ${f%%.xml}.als; done\n")
}

func gitOnMasterBranch(f *file) bool {
	head := shellExecString(f.Dir, "git", "name-rev", "--name-only", "HEAD")
	return head == "master"
}

func gitCommit(f *file) {
	gitInit(f)
	if !gitOnMasterBranch(f) {
		return
	}
	alsFile := filepath.Base(f.Name)
	xmlFile := strings.TrimSuffix(alsFile, f.Ext) + ".xml"

	shellExec(f.Dir, "sh", "-c", "cat "+alsFile+" | gunzip > "+xmlFile)
	shellExec(f.Dir, "git", "add", xmlFile)
	shellExec(f.Dir, "git", "commit", "-m", "", "--allow-empty-message")
}

func gitAmend(f *file, msg string) {
	if !gitOnMasterBranch(f) {
		return
	}
	shellExec(f.Dir, "git", "commit", "--amend", "-m", msg, "--allow-empty-message")
}

func main() {
	w, err := newWatcher(
		func(filename string) bool {
			// fmt.Println("validator got:", filename)
			return filepath.Ext(filename) == ".als"
		},
		func(filename string) {
			fmt.Println(" callback got:", filename)
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
