package main

import (
	"os"
	"path/filepath"
	"text/template"
)

type AlpineConfig struct {
	Hostname     string
	Username     string
	Password     string
	Timezone     string
	Keymap       string
	NetworkIface string
	DiskDevice   string
	Groups       []string
}

func main() {
	// Default configuration
	config := AlpineConfig{
		Hostname:     "alpinehost",
		Username:     "alpine",
		Password:     "changeme",
		Timezone:     "UTC",
		Keymap:       "us",
		NetworkIface: "eth0",
		DiskDevice:   "/dev/mmcblk0",
		Groups:       []string{"audio", "video", "netdev"},
	}

	// Get the template file path
	tmplPath := filepath.Join("templates", "answers.tmpl")

	// Parse and execute the template
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		panic(err)
	}

	err = t.Execute(os.Stdout, config)
	if err != nil {
		panic(err)
	}
}
