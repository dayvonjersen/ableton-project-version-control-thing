package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func normalizePathSeparators(path string) string {
	return strings.Replace(path, "\\", "/", -1)
}

func splitFilename(filename string) (dir, file string) {
	filename = normalizePathSeparators(filename)
	dir = normalizePathSeparators(filepath.Dir(filename))
	return dir, strings.TrimPrefix(strings.TrimPrefix(filename, dir), "/")
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

func fileGetContents(filename string) string {
	contents := new(bytes.Buffer)
	f, err := os.Open(filename)
	checkErr(err)
	_, err = io.Copy(contents, f)
	if err != io.EOF {
		checkErr(err)
	}
	checkErr(f.Close())
	return contents.String()
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
func shellExecAttach(rundir, command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Dir = rundir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Println(command, strings.Join(args, " "), ":", err)
	}
}

func shellExecSilent(rundir, command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Dir = rundir
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard
	if err := cmd.Run(); err != nil {
		log.Println(command, strings.Join(args, " "), ":", err)
	}
}
