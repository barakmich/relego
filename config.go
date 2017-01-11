package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	yaml "gopkg.in/yaml.v2"
)

var (
	version string
)

func init() {
	flag.StringVar(&version, "version", "", "version string to use for ld")
}

type Target struct {
	OS     string
	Arch   string
	Config *Config
	Opt    map[string]interface{}
}

func (t Target) String() string {
	return fmt.Sprintf("%s_v%s_%s_%s", t.Config.Name, t.Config.version, t.OS, t.Arch)
}

func (t Target) GetOpt(s string) interface{} {
	if t.Opt == nil {
		return nil
	}
	return t.Opt[s]
}

func (t *Target) SetOpt(s string, i interface{}) {
	if t.Opt == nil {
		t.Opt = make(map[string]interface{})
	}
	t.Opt[s] = i
}

type BuildTarget struct {
	OS   string   `yaml:"os"`
	Arch []string `yaml:"arch"`
}

type Config struct {
	Name      string        `yaml:"name"`
	Matrix    []BuildTarget `yaml:"matrix"`
	Mains     []string      `yaml:"mains"`
	Include   []string      `yaml:"include"`
	OutputDir string        `yaml:"outputDir"`
	LD        *LDConfig     `yaml:"ld"`
	version   string
	buildDate time.Time
}

type LDConfig struct {
	VersionPath   string `yaml:"versionPath"`
	BuildDatePath string `yaml:"buildDatePath"`
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
	err = out.FillDefaults(path)
	return &out, err
}

func (c *Config) FillDefaults(path string) error {
	if len(c.Matrix) == 0 {
		log.Printf("WARNING: No arch list provided. Using GOOS & GOARCH (%s_%s).\n", runtime.GOOS, runtime.GOARCH)
		c.Matrix = append(c.Matrix, BuildTarget{runtime.GOOS, []string{runtime.GOARCH}})
	}
	if len(c.Mains) == 0 {
		log.Printf("WARNING: No main package list provided. Using \".\".\n")
		c.Mains = append(c.Mains, ".")
	}
	if len(c.OutputDir) == 0 {
		c.OutputDir = "."
	}
	c.buildDate = time.Now()
	c.version = version
	if c.version == "" {
		c.version = c.buildDate.Format("2006-01-02-07-04")
	}
	if c.Name == "" {
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		c.Name = filepath.Base(filepath.Dir(abs))
		fmt.Println(c.Name)
	}
	return nil
}
