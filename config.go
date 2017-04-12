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
	Platform string
	Arch     string
	Config   *Config
	Opt      map[string]interface{}
}

func (t Target) IsSrc() bool {
	return t.Platform == "source" || t.Platform == "src"
}

func (t Target) String() string {
	if t.IsSrc() {
		return fmt.Sprintf("%s_%s_src", t.Config.Name, t.Config.version)
	}
	return fmt.Sprintf("%s_%s_%s_%s", t.Config.Name, t.Config.version, t.Platform, t.Arch)
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
	Platform string   `yaml:"platform"`
	Arch     []string `yaml:"arch"`
}

func (bt BuildTarget) IsSrc() bool {
	return bt.Platform == "source" || bt.Platform == "src"
}

type GlideConfig struct {
	Path        string `yaml:"glidePath"`
	GlideVC     bool   `yaml:"glide-vc"`
	GlideVCPath string `yaml:"glideVCPath"`
}

type Config struct {
	Name        string        `yaml:"name"`
	BuildMatrix []BuildTarget `yaml:"build"`
	GlideConfig *GlideConfig  `yaml:"glide,omitempty"`
	Mains       []string      `yaml:"mains"`
	Include     []string      `yaml:"include"`
	OutputDir   string        `yaml:"outputDir"`
	LD          *LDConfig     `yaml:"ld,omitempty"`
	version     string
	buildDate   time.Time
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
	if len(c.BuildMatrix) == 0 {
		log.Printf("WARNING: No build matrix provided. Using GOOS & GOARCH (%s_%s).\n", runtime.GOOS, runtime.GOARCH)
		c.BuildMatrix = append(c.BuildMatrix, BuildTarget{runtime.GOOS, []string{runtime.GOARCH}})
	}
	if len(c.Mains) == 0 {
		log.Printf("WARNING: No main package list provided. Using \".\".\n")
		c.Mains = append(c.Mains, ".")
	}
	if len(c.OutputDir) == 0 {
		c.OutputDir = "."
	}
	c.OutputDir = os.ExpandEnv(c.OutputDir)
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
	}
	return nil
}
