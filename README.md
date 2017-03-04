# go-watcher
This cli monitors your app and lib directories and rebuild when you made changes to your source files

## Installation
go get github.com/ndphu/go-watcher
go install github.com/ndphu/go-watcher

## Usage
go-watcher --help
```
GLOBAL OPTIONS:
   --work-dir value        monitoring directory (default: <your_current_dir_here>)
   --lib-dirs value        define list of directories to monitor (useful for modifying both main code and libraries' code)
   --pattern value         pattern for matching the source file (default: ".*\\.go$")
   --watch-interval value  monitoring sleep timeout in millisecond (default: 2000)
   --print-stdout          print child process's stdout
   --print-stderr          print child process's stderr
   --help, -h              show help
   --version, -v           print the version
   
```

## Authors
Phu Nguyen <ngdacphu.khtn@gmail.com>

