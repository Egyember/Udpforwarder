package main

import (
	"flag"
	"os"
	"text/template"
)

var (
	bin         = flag.String("b", "udpforward", "binary name")
	target      = flag.String("d", "dockerfile", "dockerfile name")
	templateLoc = flag.String("t", "dockerfile.temp", "location of template")
	config      = flag.String("c", "config.toml", "location of config")
)

type args struct {
	Bin    string
	Config string
}

func main() {
	flag.Parse()
	tmp, err := template.ParseFiles(*templateLoc)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(*target)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var arg args
	arg.Bin = *bin
	arg.Config = *config
	err = tmp.Execute(f, arg)
	if err != nil {
		panic(err)
	}
}
