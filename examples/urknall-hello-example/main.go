package main

import (
	"log"
	"os"

	"github.com/megamsys/urknall"
)

func main() {
	if e := provision(); e != nil {
		log.Fatal(e)
	}
}

func provision() error {
	// setup logging to stdout
	defer urknall.OpenLogger(os.Stdout).Close()

	// create a basic urknall.Template
	// executes "echo hello world" as user ubuntu on the provided host
	tpl := urknall.TemplateFunc(func(p urknall.Package) {
		p.AddCommands("run", Shell("echo hello world"))
	})

	// create provisioning target for provisioning via ssh with
	// user=ubuntu
	// host=172.16.223.142
	// password=ubuntu
	target, e := urknall.NewSshTargetWithPassword("ubuntu@172.16.223.142", "ubuntu")
	if e != nil {
		return e
	}
	return urknall.Run(target, tpl)
}
