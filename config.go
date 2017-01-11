package main

import (
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	OS          []string `yaml:"os"`
	Arch        []string `yaml:"arch"`
	Mains       []string `yaml:"mains"`
	Include     []string `yaml:"include"`
	BuildOutput string   `yaml:"output"`
	version     string
	buildDate   *time.Time
}

func ReadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	out := Config{}
	err = yaml.Unmarshal(b, &out)
	if err != nil {
		return nil, err
	}
	err = out.FillDefaults()
	return &out, err
}

func (c *Config) FillDefaults() error {
	if len(c.OS) == 0 {
		log.Printf("WARNING: No OS list provided. Using GOOS (%s).\n", runtime.GOOS)
		c.OS = append(c.OS, runtime.GOOS)
	}
	if len(c.Arch) == 0 {
		log.Printf("WARNING: No arch list provided. Using GOARCH (%s).\n", runtime.GOARCH)
		c.Arch = append(c.OS, runtime.GOARCH)
	}
	if len(c.Mains) == 0 {
		log.Printf("WARNING: No main package list provided. Using CWD.\n")
		c.Mains = append(c.OS, ".")
	}
	c.buildDate = time.Now()
	return nil
}
