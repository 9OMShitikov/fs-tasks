package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

var digitCheck = regexp.MustCompile(`^[0-9]+$`)

func isPidDir(f os.FileInfo) bool {
	return f.IsDir() && digitCheck.Match([]byte(f.Name()))
}

// adds symlinks from directory to list
func addLinksFromDir(path string, files *[]string) (err error) {
	symlinks, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}

	for _, link := range symlinks {
		resolved, err := os.Readlink(path + "/" + link.Name())
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		} else {
			*files = append(*files, resolved)
		}
	}
	return nil
}

// returns cmd and a list ow files opened by process with given pid
func retrieveInfo(pid string) (cmd string, files []string, err error) {
	cmdBytes, err := ioutil.ReadFile("/proc/" + pid +"/comm")
	cmd = string(cmdBytes[:len(cmdBytes) - 1])
	if err != nil {
		return
	}

	err = addLinksFromDir("/proc/" + pid + "/fd", &files)
	if err != nil {
		return
	}

	err = addLinksFromDir("/proc/" + pid + "/map_files", &files)
	if err != nil {
		return
	}
	return
}

func main() {
	contents, err := ioutil.ReadDir("/proc")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%18s %18s %s\n", "CMD", "PID", "NAME")
	for _, f := range contents {
		if isPidDir(f) {
			cmd, files, err := retrieveInfo(f.Name())
			if err != nil {
				log.Fatal(err)
			}

			for _, file := range(files) {
				fmt.Printf("%16s %16s %s\n", cmd, f.Name(), file)
			}
		}
	}
}
