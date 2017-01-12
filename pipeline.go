package main

import (
	"errors"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Pipe func(*Target) error

func MkWorkDir(t *Target) error {
	workDir := filepath.Join(os.ExpandEnv(t.Config.OutputDir), t.String())
	os.MkdirAll(workDir, 0755)
	t.SetOpt("workDir", workDir)
	return nil
}

func GoBuild(t *Target) error {
	wd := t.GetOpt("workDir")
	if wd == nil {
		return errors.New("no workDir")
	}
	workdir := wd.(string)
	envs := os.Environ()
	envs = append(envs, fmt.Sprintf("GOOS=%s", t.Platform))
	envs = append(envs, fmt.Sprintf("GOARCH=%s", t.Arch))
	for _, x := range t.Config.Mains {
		binname, err := resolveBinName(x)
		if err != nil {
			return err
		}
		if t.Platform == "windows" {
			binname += ".exe"
		}
		cmd := exec.Command("go", "build", "-o", filepath.Join(workdir, binname))
		cmd.Env = envs
		b, err := cmd.CombinedOutput()
		fmt.Println(string(b))
		if err != nil {
			return err
		}
	}
	return nil
}

func Cleanup(t *Target) error {
	x := t.GetOpt("workDir")
	if x == nil {
		return errors.New("no workDir to clean up?")
	}
	return os.RemoveAll(x.(string))
}

func resolveBinName(n string) (string, error) {
	if build.IsLocalImport(n) {
		p, err := filepath.Abs(n)
		if err != nil {
			return "", err
		}
		return filepath.Base(p), nil
	}
	s := strings.Split(n, "/")
	return s[len(s)-1], nil
}
