//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

const (
	name       = "gobl"
	mainBranch = "main"
)

func Build() error {
	changed, err := target.Dir("./"+name, ".")
	if os.IsNotExist(err) || (err == nil && changed) {
		args := []string{"build"}
		flags, err := buildFlags()
		if err != nil {
			return err
		}
		args = append(args, flags...)
		args = append(args, "./cmd/"+name)
		return sh.RunV("go", args...)
	}
	return nil
}

func Install() error {
	args := []string{"install"}
	flags, err := buildFlags()
	if err != nil {
		return err
	}
	args = append(args, flags...)
	args = append(args, "./cmd/"+name)
	return sh.RunV("go", args...)
}

func buildFlags() ([]string, error) {
	ldflags := []string{
		fmt.Sprintf("-X 'main.BuildDate=%s'", date()),
	}
	if v, err := version(); err != nil {
		return nil, err
	} else if v != "" {
		ldflags = append(ldflags, fmt.Sprintf("-X 'main.Version=%s'", v))
	}
	if c, err := commit(); err != nil {
		return nil, err
	} else if c != "" {
		ldflags = append(ldflags, fmt.Sprintf("-X 'main.BuildCommit=%s'", c))
	}

	out := []string{}
	if len(ldflags) > 0 {
		out = append(out, "-ldflags="+strings.Join(ldflags, " "))
	}
	return out, nil
}

func version() (string, error) {
	b, err := branch()
	if err != nil {
		return "", err
	}
	if b == mainBranch {
		if b, err = versionTag(); err != nil {
			return "", err
		}
	}
	if uncommittedChanges() {
		return fmt.Sprintf("%s-uncommitted", b), nil
	}

	return b, nil
}

func branch() (string, error) {
	return trimOutput("git", "rev-parse", "--abbrev-ref", "HEAD")
}

func versionTag() (string, error) {
	return trimOutput("git", "describe", "--exact-match", "--tags")
}

func commit() (string, error) {
	return trimOutput("git", "rev-parse", "--short", "HEAD")
}

func uncommittedChanges() bool {
	err := sh.Run("git", "diff-index", "--quiet", "HEAD", "--")
	return err != nil
}

func date() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func trimOutput(cmd string, args ...string) (string, error) {
	txt, err := sh.Output(cmd, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(txt), nil
}
