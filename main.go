/*
TODO
 - git init if not exist
 - git rebase for commit messages
/*
#!/bin/bash
for a in $(ls *.als); do
	cat $a | gunzip > $a.xml
	git add $a.xml
done
git commit
*/
package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"time"

	winfs "golang.org/x/exp/winfsnotify"
)

type file struct {
	Name string
	Time int64

	info string
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
	log.Println("watching...")
	w, err := winfs.NewWatcher()
	checkErr(err)

	checkErr(w.Watch("."))

	files := make(chan *file)
	rf := recentFiles{}

	go func() {
		for {
			<-time.After(time.Second * 2)
			log.Println("recent files:")
			for _, f := range rf[:] {
				fmt.Println(f.Name, f.Time)
			}
			fmt.Println()
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
				log.Println(f)
			}
		}
	}()
	for {
		select {
		case e := <-w.Event:
			// evtdir := strings.Replace(filepath.Dir(e.Name), "\\", "/", -1)
			// rundir := strings.Split(evtdir, "/")[0]
			// ext := filepath.Ext(file)

			files <- &file{
				Name: filepath.Base(e.Name),
				Time: time.Now().Unix(),

				info: e.String(),
			}
		case err := <-w.Error:
			checkErr(err)
		}
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
