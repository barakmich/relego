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
	workDir := filepath.Join(t.Config.OutputDir, t.String())
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
		if err != nil {
			fmt.Println(string(b))
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

func CopyFiles(t *Target) error {
	workDir := t.GetOpt("workDir")
	if workDir == nil {
		return errors.New("no workDir")
	}
	if len(t.Config.Include) == 0 {
		return nil
	}
	opts := []string{"-r"}
	opts = append(opts, t.Config.Include...)
	opts = append(opts, workDir.(string)+"/")
	cmd := exec.Command("cp", opts...)
	cmd.Env = os.Environ()
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
	}
	return err
}

func Compress(t *Target) error {
	switch t.Platform {
	case "source", "src":
		panic("unimplemented")
	case "windows":
		return compressZip(t)
	default:
		return compressTGZ(t)
	}
}

func compressTGZ(t *Target) error {
	tarpath := filepath.Join(t.Config.OutputDir, t.String()+".tar")
	cmd := exec.Command("tar", "-cvf", tarpath, t.String())
	cmd.Dir = t.Config.OutputDir
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
		return err
	}
	cmd = exec.Command("gzip", tarpath)
	cmd.Dir = t.Config.OutputDir
	b, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
		return err
	}
	return nil
}

func compressZip(t *Target) error {
	zippath := filepath.Join(t.Config.OutputDir, t.String()+".zip")
	env := []string{
		fmt.Sprintf("PWD=%s", t.Config.OutputDir),
	}
	cmd := exec.Command("zip", "-r", zippath, ".", "-i", t.String())
	cmd.Env = env
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
		return err
	}
	return nil
}

func Glide(t *Target) error {
	panic("unimplemented")
}
func SrcBuild(t *Target) error {
	panic("unimplemented")
}
