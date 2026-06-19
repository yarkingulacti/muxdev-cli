#!/usr/bin/env python3
"""Generate latest.json / manifest.json for a Goreleaser dist/ directory."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path


def build_manifest(version: str, tag: str, base_url: str, dist_dir: Path) -> dict:
    assets: dict[str, str] = {}
    for path in sorted(dist_dir.glob(f"muxdev_{version}_*")):
        name = path.name
        if name.endswith(".tar.gz"):
            key = name.removeprefix(f"muxdev_{version}_").removesuffix(".tar.gz")
        elif name.endswith(".zip"):
            key = name.removeprefix(f"muxdev_{version}_").removesuffix(".zip")
        else:
            continue
        assets[key] = name

    if not assets:
        raise SystemExit(f"no muxdev_{version}_* archives found in {dist_dir}")

    base_url = base_url.rstrip("/")
    return {
        "version": version,
        "tag": tag,
        "base_url": base_url,
        "checksums": f"{base_url}/checksums.txt",
        "assets": assets,
    }


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("version", help="SemVer without v prefix, e.g. 1.0.0")
    parser.add_argument("tag", help="Git tag, e.g. v1.0.0")
    parser.add_argument("base_url", help="Artifact base URL for this release")
    parser.add_argument("dist_dir", type=Path, help="Goreleaser dist directory")
    parser.add_argument(
        "-o",
        "--output",
        type=Path,
        help="Write JSON to this file (default: stdout)",
    )
    args = parser.parse_args()

    payload = build_manifest(args.version, args.tag, args.base_url, args.dist_dir)
    text = json.dumps(payload, indent=2) + "\n"

    if args.output:
        args.output.write_text(text, encoding="utf-8")
    else:
        sys.stdout.write(text)


if __name__ == "__main__":
    main()
