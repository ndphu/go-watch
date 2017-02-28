package main

import (
	"crypto/md5"
	"fmt"
	"github.com/urfave/cli"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"
)

var (
	FileHash         = ""
	SleepDelay int64 = 500

	Process *os.Process = nil
)

func ComputeMd5(filePath string) ([]byte, error) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}

func KillExistingProcess() {
	if Process != nil {
		log.Println("Killing existing child project")
		Process.Kill()
	}
}

func OnFileChange() {
	KillExistingProcess()

	out, err := exec.Command("go", "build", "main.go").CombinedOutput()
	if err != nil {
		log.Println("Build failed", err)
		log.Printf("Build output:", string(out))
		return
	} else {
		log.Println("Build successful")
	}

	cmd := exec.Command("./main")
	err = cmd.Start()
	if err != nil {
		log.Println("Execution failed", err)
	} else {
		log.Println("Child process PID", cmd.Process.Pid)
		Process = cmd.Process
		log.Println("Done execution new binary")
	}
}

func main() {
	defer KillExistingProcess()

	app := cli.NewApp()

	app.Name = "Go Watcher"
	app.Usage = "Monitor go source files and auto rebuild & reload the app when any source file is changed"
	app.Author = "Phu Nguyen <ngdacphu.khtn@gmail.com>"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "workdir",
			Value: "./",
			Usage: "monitoring directory",
		},
		cli.StringFlag{
			Name:  "pattern",
			Value: ".*\\.go$",
			Usage: "pattern for matching the source file",
		},
	}

	app.Action = func(c *cli.Context) error {
		WorkDir := c.String("workdir")
		Pattern := c.String("pattern")
		log.Printf("Using working directory %s\n", WorkDir)
		os.Chdir(WorkDir)
		cacheMap := make(map[string]string)
		for {
			changed := false
			// TODO handle file removal
			sourceFiles := ListSourceFile(WorkDir, Pattern)
			for _, f := range sourceFiles {
				h, _ := ComputeMd5(f)
				newHash := fmt.Sprintf("%x", h)
				oldHash, exists := cacheMap[f]
				if !exists || strings.Compare(oldHash, newHash) != 0 {
					changed = true
					cacheMap[f] = newHash
				}
			}

			if changed {
				OnFileChange()
			}
			time.Sleep(500)
		}

		return nil
	}

	app.Run(os.Args)

}

func ListSourceFile(baseDir string, pattern string) []string {
	result := make([]string, 0)
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return make([]string, 0)
	}
	for _, e := range files {
		if e.IsDir() {
			for _, _e := range ListSourceFile(path.Join(baseDir, e.Name()), pattern) {
				result = append(result, _e)
			}
		} else {
			match, _ := regexp.MatchString(pattern, e.Name())
			if match {
				result = append(result, path.Join(baseDir, e.Name()))
			}
		}
	}
	return result
}
