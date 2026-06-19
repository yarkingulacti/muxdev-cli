package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/yarkingulacti/muxdev-cli/internal/version"
)

func downloadToFile(ctx context.Context, asset Asset, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return err
	}
	setRequestHeaders(req, version.Short())

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download asset: http %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}
