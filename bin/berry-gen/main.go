package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const template = `package main

import (
  . "github.com/zyxar/berry/core"
)

func main() {

}

`

const lib_template = `package %s

import (
  . "github.com/zyxar/berry/core"
)

`

var mklib bool

func main() {
	flag.BoolVar(&mklib, "lib", false, "make a library package")
	flag.Parse()
	if flag.NArg() == 0 || flag.Arg(0) == "" {
		fmt.Println("No package name provided.\n")
		fmt.Println("usage: berry-gen [option] {PKG_NAME}")
		flag.PrintDefaults()
		os.Exit(1)
	}
	package_name := flag.Arg(0)
	file_name := "main.go"
	if flag.NArg() > 1 || flag.Arg(1) != "" {
		file_name = flag.Arg(1)
		if !strings.HasSuffix(file_name, ".go") {
			file_name += ".go"
		}
	} else if mklib {
		file_name = package_name + ".go"
	}
	var err error
	defer func() {
		if err != nil {
			fmt.Println(err)
		}
	}()
	os.Mkdir(package_name, 0755)
	if err = os.Chdir(package_name); err != nil {
		return
	}
	if _, err = os.Stat(file_name); err == nil {
		err = os.ErrExist
		return
	}
	file, err := os.Create(file_name)
	if err != nil {
		return
	}
	if mklib {
		_, err = io.WriteString(file, fmt.Sprintf(lib_template, package_name))
		file.Close()
	} else {
		_, err = io.WriteString(file, template)
		file.Close()
		if file, err = os.Create(".gitignore"); err != nil {
			return
		}
		_, err = io.WriteString(file, package_name+"\n")
	}
	return
}
