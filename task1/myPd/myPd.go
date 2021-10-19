package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

func isPidDir(f os.FileInfo) bool {
	var digitCheck = regexp.MustCompile(`^[0-9]+$`)
	return f.IsDir() && digitCheck.Match([]byte(f.Name()))
}

// returns cmd, vsize and rss of process with given pid
func retrieveInfo(pid string, rexp *regexp.Regexp) (string, string, string, error) {
	processStats, err := ioutil.ReadFile("/proc/" + pid +"/stat")
	if err != nil {
		return "", "", "", err
	}
	statsString := string(processStats)

	res := rexp.FindStringSubmatch(statsString)
	if len(res) == 0 {
		return "", "", "", errors.New("stat file doesn't match format")
	}

	cmd, err := ioutil.ReadFile("/proc/" + pid +"/comm")

	if err != nil {
		return "", "", "", err
	}

	return string(cmd[: len(cmd) - 1]), res[1], res[2], nil
}

func main() {
	contents, err := ioutil.ReadDir("/proc")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%10s %16s %16s %s\n", "PID", "RSS", "VSIZE", "CMD")
	rexp := regexp.MustCompile(`\d \(.*\) \S -?\d+ ` + // pid, comm, state, ppid
		`-?\d+ -?\d+ -?\d+ -?\d+ ` + // pgrp, session, tty_nr, tpgid
		`\d+ \d+ \d+ \d+ ` + // flags, minflt, cminflt, majflt
		`\d+ \d+ \d+ -?\d+ ` + // cmajflt, utime, stime, cutime
		`-?\d+ -?\d+ -?\d+ -?\d+ ` + // cstime, priority, nice, num_threads
		`-?\d+ \d+ (\d+) (-?\d+)`) //itrealvalue, starttime, vsize, rss

	for _, f := range contents {
		if isPidDir(f) {
			cmd, vsize, rss, err := retrieveInfo(f.Name(), rexp)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%10s %16s %16s %s\n", f.Name(), rss, vsize, cmd)
		}
	}
}
