package sver

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func git(args ...string) (string, error) {
	cmd := exec.Command(gitBinary, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "unexpected result from git; output: \n%s\n", string(out))
	}

	return strings.TrimSpace(string(out)), nil
}

func verifyGit() error {
	_, err := exec.LookPath(gitBinary)
	if err != nil {
		return errors.New("git not found in your PATH; please install it")
	}

	cmd := exec.Command(gitBinary, "rev-parse", "--is-inside-work-tree")
	stdErrBuf := new(bytes.Buffer)
	cmd.Stderr = stdErrBuf
	err = cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "could not determine if the current directory is a git working tree: %s", stdErrBuf.String())
	}

	return nil
}
