package main

import (
	"flag"
	"fmt"
	"os"

	. "github.com/donnie4w/tfdoc"
)

func main() {
	dir := ""
	tofile := "newfile.thrift"
	namespace := ""
	flag.StringVar(&dir, "dir", "", "")
	flag.StringVar(&tofile, "tofile", "newfile.thrift", "")
	flag.StringVar(&namespace, "namespace", "", "namespace")
	flag.Parse()
	if dir == "" {
		fmt.Println("dir is empty")
		os.Exit(1)
	}

	WalkDir(dir, "thrift")
	CreateNewThrifFile(tofile, namespace)
}
