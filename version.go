package main

import (
	"fmt"
	"runtime"
	"time"
)

// package global var, value set by linker using ldflag -X
var (
	ver    string //nolint:gochecknoglobals
	date   string //nolint:gochecknoglobals
	commit string //nolint:gochecknoglobals
)

// Info - version info.
type Info struct {
	Version string
	Date    string
	Commit  string
}

// GetInfo - get version stamp information.
func GetInfo() Info {
	if ver == "" {
		ver = "0.0.0"
	}

	if date == "" {
		date = time.Now().Format(time.RFC3339)
	}

	if commit == "" {
		commit = "????????"
	}

	return Info{
		Version: ver,
		Date:    date,
		Commit:  commit,
	}
}

// String() -- return version info string.
func (vi Info) String() string {
	return fmt.Sprintf("%s g%s %s-%s [%s]",
		vi.Version,
		vi.Commit,
		runtime.GOOS,
		runtime.GOARCH,
		vi.Date,
	)
}
