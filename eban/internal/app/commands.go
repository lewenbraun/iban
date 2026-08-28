package app

import (
	"fmt"
)

var commands = map[string]func([]string) error{
	"toggle": cmdToggle,
	"start":  cmdStart,
	"stop":   cmdStop,
	"status": cmdStatus,
}

func cmdToggle(args []string) error {
	svc := newService()
	if svc.Active() {
		return svc.Stop("", false)
	}
	return cmdStart(args)
}

func cmdStart(args []string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}
	return newService().Start(opts.mode(), opts.lang, opts.timeout)
}

func cmdStop(args []string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}
	return newService().Stop(opts.lang, opts.paste)
}

func cmdStatus([]string) error {
	fmt.Println(status())
	return nil
}
