package version

import (
	"encoding/json"
	"fmt"
	"runtime"
)

var (
	Version string
	Commit  string
	Date    string
)

type Build struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
	GoOS      string `json:"go_os"`
	GoArch    string `json:"go_arch"`
}

func (b Build) String() string {
	return fmt.Sprintf(
		"Build{Version: %s, Commit: %s, Date: %s, GoVersion: %s, GoOS: %s, GoArch: %s}",
		b.Version,
		b.Commit,
		b.Date,
		b.GoVersion,
		b.GoOS,
		b.GoArch,
	)
}

func (b Build) JSON() string {
	m, err := json.Marshal(b)
	if err != nil {
		return fmt.Sprintf("json marshal error: %v", err)
	}

	return string(m)
}

func (b Build) Pretty() string {
	return fmt.Sprintf(
		"Version:\t%s\nCommit:\t\t%s\nDate:\t\t%s\n\nGo Version:\t%s\nGo OS:\t\t%s\nGo Arch:\t%s",
		b.Version,
		b.Commit,
		b.Date,
		b.GoVersion,
		b.GoOS,
		b.GoArch,
	)
}

func GetBuild() Build {
	return Build{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: runtime.Version(),
		GoOS:      runtime.GOOS,
		GoArch:    runtime.GOARCH,
	}
}

func init() {
	if Version == "" {
		Version = "dev"
	}

	if Commit == "" {
		Commit = "unknown"
	}

	if Date == "" {
		Date = "unknown"
	}
}
