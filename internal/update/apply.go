package update

import (
	"context"
	"os"
	"path/filepath"
)

type ApplyOptions struct {
	Release   Release
	AssetName string
}

func Apply(ctx context.Context, opts ApplyOptions) error {
	exe, err := CurrentExecutable()
	if err != nil {
		return err
	}

	asset, err := FindAsset(opts.Release, opts.AssetName)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "muxdev-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, opts.AssetName)
	binaryPath := filepath.Join(tmpDir, "muxdev-new")

	if err := downloadToFile(ctx, asset, archivePath); err != nil {
		return err
	}
	checksums, err := DownloadChecksums(ctx, opts.Release)
	if err != nil {
		return err
	}
	if err := VerifyChecksum(checksums, opts.AssetName, archivePath); err != nil {
		return err
	}
	if err := extractBinary(archivePath, opts.AssetName, binaryPath); err != nil {
		return err
	}

	return replaceExecutable(exe, binaryPath)
}

func replaceExecutable(target, source string) error {
	return applyReplace(target, source)
}
