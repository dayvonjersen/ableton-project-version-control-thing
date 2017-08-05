/*
	TODO:
		- "UI" improvements
		- git tags ???
*/
package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	winfs "golang.org/x/exp/winfsnotify"
)

func normalizePathSeparators(path string) string {
	return strings.Replace(path, "\\", "/", -1)
}

func fileExists(filename string) bool {
	f, err := os.Open(filename)
	if os.IsNotExist(err) {
		return false
	}
	checkErr(err)
	checkErr(f.Close())
	return true
}

func dirExists(path string) bool {
	finfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	checkErr(err)
	return finfo.IsDir()
}

func isDir(filename string) bool {
	finfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	checkErr(err)
	return finfo.IsDir()
}

func filePutContents(filename, contents string) {
	f, err := os.Create(filename)
	checkErr(err)
	_, err = io.WriteString(f, contents)
	checkErr(err)
	checkErr(f.Close())
}

func shellExec(rundir, command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Dir = rundir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Println(command, strings.Join(args, " "), ":", err)
	}
}

func shellExecString(rundir, command string, args ...string) string {
	cmd := exec.Command(command, args...)
	cmd.Dir = rundir
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Println(command, strings.Join(args, " "), ":", err)
	}
	return strings.TrimSpace(string(out))
}

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

func main() {
	w, err := winfs.NewWatcher()
	checkErr(err)

	watchDir := "."
	watchPaths := []string{}
	watchPath := func(name string) {
		name = normalizePathSeparators(name)
		for _, path := range watchPaths {
			if path == name {
				return
			}
		}
		log.Println("watching", name)
		watchPaths = append(watchPaths, name)
		checkErr(w.AddWatch(name, winfs.FS_MODIFY|winfs.FS_CREATE|winfs.FS_MOVED_TO))
	}
	watchPath(watchDir)

	filepath.Walk(watchDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !strings.Contains(path, ".git") {
			if !dirExists(path + "/.git") {
				if matches, err := filepath.Glob("*.als"); err == nil && matches != nil {
					for _, m := range matches {
						f := &file{
							Name: path + "/" + m,
							Dir:  path,
							Ext:  ".als",
							Time: time.Now().Unix(),
						}
						gitCommit(f)
					}
				}
			}
			watchPath(path)
		}
		return err
	})

	files := make(chan *file)
	rf := recentFiles{}

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if len(rf) > 0 {
				gitAmend(rf[0], scanner.Text())
			}
		}
	}()

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
			} else if isDir(name) && !strings.Contains(name, ".git") {
				watchPath(name)
			}
		case err := <-w.Error:
			checkErr(err)
		}
	}
}

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
