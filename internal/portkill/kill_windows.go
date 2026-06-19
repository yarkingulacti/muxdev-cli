//go:build windows

package portkill

import "fmt"

func KillPort(port int) (int, error) {
	return 0, fmt.Errorf("port kill is not supported on windows yet")
}

func PIDsOnPort(port int) ([]int, error) {
	return nil, fmt.Errorf("port lookup is not supported on windows yet")
}
