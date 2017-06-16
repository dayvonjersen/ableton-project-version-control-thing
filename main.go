/*
TODO
 - git rebase for commit messages
	- watch a specially named text file ??? (hammers, nails, etc...)
*/
package main

import (
	"fmt"
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
	// fmt.Println("debug:", filename)
	f, err := os.Open(filename)
	if os.IsNotExist(err) {
		return false
	}
	checkErr(err)
	checkErr(f.Close())
	return true
}

func isDir(filename string) bool {
	finfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	checkErr(err)
	return finfo.IsDir()
}

type file struct {
	Name, Dir, Ext string
	Time           int64
}

func gitInit(f *file) {
	gitDir := f.Dir + "/.git"
	// fmt.Println("debug:", gitDir)
	if !fileExists(gitDir) {
		{
			cmd := exec.Command("git", "init")
			cmd.Dir = f.Dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Println("error:", err)
			}
		}
		{
			f, err := os.OpenFile(gitDir+"/info/exclude", os.O_RDWR|os.O_CREATE, 0755)
			checkErr(err)
			io.WriteString(f, "*\n!*.xml\n")
			checkErr(f.Close())
		}
		{
			f, err := os.OpenFile(gitDir+"/hooks/post-checkout", os.O_RDWR|os.O_CREATE, 0755)
			checkErr(err)
			io.WriteString(f, "#!/bin/bash\nfor f in `ls --color=never *.xml`; do cat $f | gzip > ${f%%.xml}.als; done\n")
			checkErr(f.Close())
		}
	}
}

func gitCommit(f *file) {
	gitInit(f)
	alsFile := filepath.Base(f.Name)
	xmlFile := strings.TrimSuffix(alsFile, f.Ext) + ".xml"
	{
		cmd := exec.Command("sh", "-c", "cat "+alsFile+" | gunzip > "+xmlFile)
		cmd.Dir = f.Dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Println("error:", err)
		}
	}
	{
		cmd := exec.Command("git", "add", xmlFile)
		cmd.Dir = f.Dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Println("error:", err)
		}
	}
	{
		cmd := exec.Command("git", "commit", "-m", "", "--allow-empty-message")
		cmd.Dir = f.Dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Println("error:", err)
		}
	}
}

type recentFiles []*file

func (rf recentFiles) Len() int {
	return len(rf)
}
func (rf recentFiles) Less(i, j int) bool {
	if rf[i].Name == rf[j].Name {
		return rf[i].Time > rf[j].Time
	}
	return rf[i].Name < rf[j].Name
}
func (rf recentFiles) Swap(i, j int) {
	rf[i], rf[j] = rf[j], rf[i]
}
func (rf recentFiles) Find(f *file) (int, bool) {
	i := sort.Search(len(rf), func(i int) bool {
		return rf[i].Name == f.Name
	})
	return i, i < len(rf) && rf[i].Name == f.Name
}

func main() {
	w, err := winfs.NewWatcher()
	checkErr(err)

	watchDir := "./test"
	watchPaths := []string{watchDir}
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

	filepath.Walk(watchDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !strings.Contains(path, ".git") {
			watchPath(path)
		}
		return err
	})

	files := make(chan *file)
	rf := recentFiles{}

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
			// fmt.Println("debug:", e)
			name := normalizePathSeparators(e.Name)
			dir := normalizePathSeparators(filepath.Dir(name))
			// fmt.Println("debug: name:", name)
			// fmt.Println("debug: dir:", dir)
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
