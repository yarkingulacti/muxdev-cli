//go:build windows

package portkill

import "fmt"

func ProcessOnPort(port int) (Process, error) {
	return Process{}, fmt.Errorf("attach is not supported on windows yet")
}
