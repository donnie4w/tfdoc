package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	. "github.com/donnie4w/tfdoc"
)

func main() {
	dir, _ := os.Getwd()
	tofile := "newfile.thrift"
	java := ""
	_go := ""
	cpp := ""
	php := ""
	py := ""
	flag.StringVar(&dir, "dir", "", "")
	flag.StringVar(&tofile, "tofile", "newfile.thrift", "")
	flag.StringVar(&java, "java", "", "java namespace")
	flag.StringVar(&_go, "go", "", "go namespace")
	flag.StringVar(&cpp, "cpp", "", "c++ namespace")
	flag.StringVar(&php, "php", "", "php namespace")
	flag.StringVar(&py, "py", "", "python namespace")
	flag.Parse()
	if dir == "" {
		fmt.Println("dir is empty")
		os.Exit(1)
	}
	WalkDir(dir, "thrift")
	if java != "" {
		java = fmt.Sprint("namespace java ", java)
	}
	if _go != "" {
		_go = fmt.Sprint("namespace go ", _go)
	}
	if cpp != "" {
		cpp = fmt.Sprint("namespace cpp ", cpp)
	}
	if php != "" {
		php = fmt.Sprint("namespace php ", php)
	}
	if py != "" {
		py = fmt.Sprint("namespace py ", py)
	}
	CreateNewThrifFile(tofile, joinEndNL(java, _go, cpp, php, py))
}

func joinEndNL(ss ...string) string {
	var buf bytes.Buffer
	for _, s := range ss {
		if s != "" {
			buf.WriteString(s)
			buf.WriteString("\n")
		}
	}
	return buf.String()
}
