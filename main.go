package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

var (
	WatchFile        = "main.go"
	WorkDir          = "./"
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
	WorkDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get current dir", err)
	}
	log.Printf("Watching file %s in %s\n", WatchFile, WorkDir)
	for {
		if b, err := ComputeMd5(path.Join(WorkDir, WatchFile)); err != nil {
			log.Printf("Err: %v\n", err)
		} else {
			newHash := fmt.Sprintf("%x", b)
			if strings.Compare(newHash, FileHash) != 0 {
				log.Printf("Change detected\n")
				OnFileChange()
			}
			FileHash = newHash
		}
		time.Sleep(100 * time.Millisecond)
	}

}
