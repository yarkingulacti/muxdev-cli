//go:build windows

package portkill

import (
	"context"
	"fmt"
)

func AttachProcess(ctx context.Context, pid int, onLine LineHandler) error {
	return fmt.Errorf("attach is not supported on windows yet")
}
