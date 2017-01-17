package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	pb "gopkg.in/cheggaaa/pb.v1"
)

var (
	config string

	prePipeline = []PrePipe{
		Glide,
	}

	pipeline = []Pipe{
		MkWorkDir,
		GoBuild,
		CopyFiles,
		Compress,
		Cleanup,
	}
	srcPipeline = []Pipe{
		MkWorkDir,
		SrcBuild,
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

	for _, t := range cfg.BuildMatrix {
		if t.IsSrc() {
			matrix = append(matrix, &Target{
				Platform: "src",
				Arch:     "src",
				Config:   cfg,
			})
			continue
		}
		for _, a := range t.Arch {
			matrix = append(matrix, &Target{
				Platform: t.Platform,
				Arch:     a,
				Config:   cfg,
			})
		}
	}

	// TODO(barakmich): parallel builds
	var wg sync.WaitGroup
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range prePipeline {
		err := f(cfg)
		if err != nil {
			log.Fatal(err)
		}
	}

	var bars []*pb.ProgressBar

	start := make(chan bool)
	for _, t := range matrix {
		wg.Add(1)
		kind := fmt.Sprintf("%s/%s", t.Platform, t.Arch)
		p := pipeline
		if t.IsSrc() {
			kind = "src"
			p = srcPipeline
		}
		bar := pb.New(len(p)).Prefix(kind)
		bar.ShowCounters = false
		bar.BarStart = " ["
		bars = append(bars, bar)
		go func(t *Target, bar *pb.ProgressBar, start chan bool) {
			defer wg.Done()
			<-start
			for _, f := range p {
				err := f(t)
				if err != nil {
					log.Fatal(err)
				}
				bar.Increment()
			}
		}(t, bar, start)
	}
	pool, err := pb.StartPool(bars...)
	if err != nil {
		log.Fatalln(err)
	}
	close(start)
	wg.Wait()
	pool.Stop()
	fmt.Println("done")

}
