package main

import (
	"bufio"
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
	WorkDir       string      = ""
	LibDirs       []string    = []string{}
	Pattern       string      = ""
	PrintStdout   bool        = false
	PrintStderr   bool        = false
	Process       *os.Process = nil
	WatchInterval int
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
		log.Printf("Killing child process [%d]\n", Process.Pid)
		Process.Kill()
	}
}

func StreamMonitor(reader io.ReadCloser, pid int, streamName string) {
	buff := bufio.NewScanner(reader)
	for buff.Scan() {
		log.Printf("[%d][%s] %s\n", pid, streamName, buff.Text())
	}
}

func GoGetLibDirs() error {
	// need to invalidate the cache for GoSublime

	for _, libDir := range LibDirs {
		os.Chdir(libDir)
		out, err := exec.Command("go", "get").CombinedOutput()
		if err != nil {
			log.Println("go get failed in", libDir, "with error", string(out))
			return err
		}
	}
	return nil
}

func OnFileChange() {
	KillExistingProcess()

	if err := GoGetLibDirs(); err != nil {
		log.Println("Failed to execute go get for libs")
		return
	}

	os.Chdir(WorkDir)

	out, err := exec.Command("go", "build", "main.go").CombinedOutput()
	if err != nil {
		log.Println("Build failed", err)
		log.Printf("Build output:", string(out))
		return
	} else {
		log.Println("Build successful")
	}

	cmd := exec.Command("./main")
	stdout, err := cmd.StdoutPipe()
	stderr, err := cmd.StderrPipe()

	if err != nil {
		log.Println("Cannot get StdoutPipe", err.Error())
	}
	err = cmd.Start()
	//defer cmd.Wait()
	if err != nil {
		log.Println("Execution failed", err)
	} else {
		log.Println("Child process PID", cmd.Process.Pid)
		Process = cmd.Process
		log.Println("Done execution new binary")
		if PrintStdout {
			go StreamMonitor(stdout, cmd.Process.Pid, "stdout")
		}
		if PrintStderr {
			go StreamMonitor(stderr, cmd.Process.Pid, "stderr")
		}

	}
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

func CheckChangeInFiles(files []string, cacheMap map[string]string) bool {
	changed := false
	for _, f := range files {
		h, _ := ComputeMd5(f)
		newHash := fmt.Sprintf("%x", h)
		oldHash, exists := cacheMap[f]
		if !exists || strings.Compare(oldHash, newHash) != 0 {
			changed = true
			cacheMap[f] = newHash
		}
	}
	return changed
}

func main() {
	defer KillExistingProcess()
	currentExecDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	app := cli.NewApp()

	app.Name = "Go Watcher"
	app.Usage = "Monitor go source files and auto rebuild & reload the app when any source file is changed"
	app.Author = "Phu Nguyen <ngdacphu.khtn@gmail.com>"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "work-dir",
			Value: currentExecDir,
			Usage: "monitoring directory",
		},
		cli.StringSliceFlag{
			Name:  "lib-dirs",
			Value: &cli.StringSlice{},
			Usage: "define list of directories to monitor (useful for modifying both main code and libraries' code)",
		},
		cli.StringFlag{
			Name:  "pattern",
			Value: ".*\\.go$",
			Usage: "pattern for matching the source file",
		},
		cli.IntFlag{
			Name:  "watch-interval",
			Value: 2000,
			Usage: "monitoring sleep timeout in millisecond",
		},
		cli.BoolTFlag{
			Name:  "print-stdout",
			Usage: "print child process's stdout",
		},
		cli.BoolTFlag{
			Name:  "print-stderr",
			Usage: "print child process's stderr",
		},
	}

	app.Action = func(c *cli.Context) error {
		WorkDir = c.String("work-dir")
		LibDirs = c.StringSlice("lib-dirs")
		Pattern = c.String("pattern")
		WatchInterval = c.Int("watch-interval")
		PrintStdout = c.BoolT("print-stdout")
		PrintStderr = c.BoolT("print-stderr")
		log.Printf("Using working directory %s\n", WorkDir)
		log.Println("Using lib dir:")
		for _, lib := range LibDirs {
			log.Println(lib)
		}
		os.Chdir(WorkDir)
		cacheMap := make(map[string]string)
		for {
			// TODO: handle file removal
			sourceFiles := ListSourceFile(WorkDir, Pattern)
			appCodeChanged := CheckChangeInFiles(sourceFiles, cacheMap)
			if appCodeChanged {
				log.Println("Change detected in application code")
			}
			libCodeChanged := false
			for _, libDir := range LibDirs {
				libFiles := ListSourceFile(libDir, Pattern)
				if CheckChangeInFiles(libFiles, cacheMap) && !libCodeChanged {
					libCodeChanged = true
					log.Println("Change detected in libs")
				}
			}

			if appCodeChanged || libCodeChanged {
				OnFileChange()
			}

			time.Sleep(time.Duration(WatchInterval) * time.Millisecond)
		}

		return nil
	}

	app.Run(os.Args)

}
