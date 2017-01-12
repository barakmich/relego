package main

import (
	"flag"
	"log"
)

var (
	config   string
	pipeline = []Pipe{
		MkWorkDir,
		GoBuild,
		CopyFiles,
		Compress,
		Cleanup,
	}
	srcPipeline = []Pipe{
		MkWorkDir,
		Glide,
		SrcBuild,
		Compress,
		Cleanup,
	}
)

func init() {
	flag.StringVar(&config, "config", ".relego.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()
	cfg, err := ReadConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	var matrix []*Target

	for _, t := range cfg.Matrix {
		for _, a := range t.Arch {
			matrix = append(matrix, &Target{
				Platform: t.Platform,
				Arch:     a,
				Config:   cfg,
			})
		}
	}

	// TODO(barakmich): parallel builds
	for _, t := range matrix {
		for _, f := range pipeline {
			err := f(t)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

}
