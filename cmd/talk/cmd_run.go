package main

import (
	"os"

	"github.com/robotalks/talk/core/engine"
)

// RunCommand implements robotalk run
type RunCommand struct {
	URL         string
	ModulesDir  []string `n:"modules-dir"`
	LoadModules bool     `n:"load-modules"`
	Quiet       bool
	Spec        string
}

// Execute implements Executable
func (c *RunCommand) Execute(args []string) error {
	os.Setenv("MQHUB_URL", c.URL)
	if c.LoadModules {
		loadModules(c.ModulesDir)
	}
	return engine.Run(c.URL, c.Spec)
}
