//go:build !((windows && amd64) || (darwin && (amd64 || arm64)) || (linux && amd64))

package system

import (
	"fmt"
	"runtime"
)

func checkSupported() error {
	return fmt.Errorf(
		"unsupported platform: %s/%s - supported platforms are: windows/amd64, darwin/amd64, darwin/arm64, linux/amd64",
		runtime.GOOS,
		runtime.GOARCH,
	)
}
