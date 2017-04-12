package main

import (
	"errors"
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Pipe func(*Target) error
type PrePipe func(*Config) error

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
		cmds := []string{
			"build", "-o", filepath.Join(workdir, binname),
		}
		if ld := t.Config.LD; ld != nil {
			cmds = append(cmds, "-ldflags")
			flags := ""
			if ld.VersionPath != "" {
				flags += fmt.Sprintf("-X %s=%s ", ld.VersionPath, t.Config.version)
			}
			if ld.BuildDatePath != "" {
				flags += fmt.Sprintf("-X %s=%s", ld.BuildDatePath, t.Config.buildDate.Format(time.RFC3339))
			}
			cmds = append(cmds, flags)
		}
		cmds = append(cmds, x)
		cmd := exec.Command("go", cmds...)
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
	// TODO(barakmich): mkdir
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
	cmd := exec.Command("zip", "-r", zippath, ".", "-i", t.String()+"/*")
	cmd.Dir = t.Config.OutputDir
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
		return err
	}
	return nil
}

func Glide(cfg *Config) error {
	if cfg.GlideConfig == nil {
		return nil
	}
	gc := cfg.GlideConfig
	run := "glide"
	if gc.Path != "" {
		run = gc.Path
	}
	cmd := exec.Command(run, "install", "--strip-vcs", "--strip-vendor", "--update-vendored", "--delete")
	cmd.Env = os.Environ()
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
		return err
	}
	if gc.GlideVC {
		run := "glide-vc"
		if gc.GlideVCPath != "" {
			run = gc.GlideVCPath
		}
		cmd := exec.Command(run, "--only-code", "--no-tests")
		cmd.Env = os.Environ()
		b, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(b))
			return err
		}

	}
	return nil

}

func SrcBuild(t *Target) error {
	wd := t.GetOpt("workDir")
	if wd == nil {
		return errors.New("no workDir")
	}

	// Collect
	tempfile := filepath.Join(wd.(string), "filelist")
	f, err := os.Create(tempfile)
	if err != nil {
		return err
	}
	if _, err := os.Stat("vendor"); err == nil {
		cmd := exec.Command("find", "vendor", "-name", ".git", "-prune", "-o", "(", "-type", "f", "-print", ")")
		cmd.Env = os.Environ()
		cmd.Stdout = f
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Printf("couldn't run `find`: %s\n", err)
			f.Close()
			return err
		}
	}
	cmd := exec.Command("git", "ls-files")
	cmd.Env = os.Environ()
	cmd.Stdout = f
	err = cmd.Run()
	if err != nil {
		fmt.Printf("couldn't run `git ls-files`: %s\n", err)
		f.Close()
		return err
	}
	f.Close()

	// Uniq
	fl := filepath.Join(wd.(string), "files")
	f, err = os.Create(fl)
	if err != nil {
		return err
	}
	cmd = exec.Command("uniq", tempfile)
	cmd.Env = os.Environ()
	cmd.Stdout = f
	err = cmd.Run()
	if err != nil {
		fmt.Printf("couldn't run `uniq`: %s\n", err)
		f.Close()
		return err
	}
	f.Close()

	// Tar
	tarpath := filepath.Join(t.Config.OutputDir, t.String()+".tar.gz")
	cmd = exec.Command("tar", "-cvzf", tarpath, "-T", fl, "--transform", fmt.Sprintf("s,^,%s/,", t.String()))
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
		return err
	}
	return nil
}
