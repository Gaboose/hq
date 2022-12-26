package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Gaboose/hq"
)

func mainErr() error {
	queryRaw := flag.Arg(0)
	if queryRaw == "" {
		_, err := io.Copy(os.Stdout, os.Stdin)
		return err
	}

	out, err := hq.Exec(queryRaw, os.Stdin)
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}

func main() {
	flag.Parse()

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println("Usage: cat some.html | hq [query]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err := mainErr(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
