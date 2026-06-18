# Local development

## Gereksinimler

- Go 1.23+ ([go.mod](../go.mod) sürümüne bak)
- Git

## Başlangıç

```bash
git clone https://github.com/yarkingulacti/muxdev-cli.git
cd muxdev-cli
go test ./...
```

## Çalıştırma

En pratik yol — her kod değişikliğinde otomatik derlenir:

```bash
./bin/muxdev --version
```

Daha hızlı tekrar çalıştırma için binary build:

```bash
go build -o muxdev ./cmd/muxdev
./muxdev --version
```

## Deneme komutları

Repo içindeki örnek config ile:

```bash
./bin/muxdev --list --config testdata/muxdev.yaml
./bin/muxdev --no-interactive --config testdata/muxdev.yaml
./bin/muxdev --config testdata/muxdev.yaml          # TUI
./bin/muxdev init                                   # config sihirbazı
```

Kendi projen için:

```bash
cd /path/to/project
/path/to/muxdev-cli/bin/muxdev init
/path/to/muxdev-cli/bin/muxdev
```

## Test

```bash
go test ./...
```

## Branch akışı

```bash
git checkout dev
git pull
git checkout -b feature/my-change
# ... commit ...
git push -u origin feature/my-change
# PR hedefi: dev
```

Detay: [git-workflow.md](git-workflow.md)

## Notlar

- Local build `muxdev version` → `dev (local build)` gösterir
- TUI ve `init`/`configure` gerçek terminal (TTY) ister
- CI/pipe testleri için `--no-interactive` kullan
