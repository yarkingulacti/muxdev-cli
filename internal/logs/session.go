package logs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yarkingulacti/muxdev-cli/internal/platform"
)

const (
	metaFilename = "meta.json"
	logFilename  = "session.log"
)

type Meta struct {
	ID         string     `json:"id"`
	StartedAt  time.Time  `json:"started_at"`
	EndedAt    *time.Time `json:"ended_at,omitempty"`
	WorkDir    string     `json:"work_dir"`
	ConfigPath string     `json:"config_path"`
	ServiceIDs []string   `json:"service_ids"`
	Runtime    string     `json:"runtime"`
	ExitError  string     `json:"exit_error,omitempty"`
}

type Session struct {
	Meta Meta
	Dir  string
}

type Writer struct {
	meta Meta
	dir  string
	file *os.File
	buf  *bufio.Writer
}

func StartSession(workDir, configPath string, serviceIDs []string, runtime string) (*Writer, error) {
	root, err := platform.SessionsDir()
	if err != nil {
		return nil, err
	}

	workDir, err = filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve work dir: %w", err)
	}
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}

	id, dir, err := allocateSessionDir(root)
	if err != nil {
		return nil, err
	}

	meta := Meta{
		ID:         id,
		StartedAt:  time.Now().UTC(),
		WorkDir:    workDir,
		ConfigPath: configPath,
		ServiceIDs: append([]string(nil), serviceIDs...),
		Runtime:    runtime,
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create session dir: %w", err)
	}
	if err := writeMeta(dir, meta); err != nil {
		return nil, err
	}

	logPath := filepath.Join(dir, logFilename)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open session log: %w", err)
	}

	return &Writer{
		meta: meta,
		dir:  dir,
		file: file,
		buf:  bufio.NewWriter(file),
	}, nil
}

func (w *Writer) Append(label, text string) error {
	if w == nil || w.buf == nil {
		return nil
	}
	if _, err := fmt.Fprintf(w.buf, "[%s] %s\n", label, text); err != nil {
		return err
	}
	return w.buf.Flush()
}

func (w *Writer) Finish(runErr error) error {
	if w == nil {
		return nil
	}

	ended := time.Now().UTC()
	w.meta.EndedAt = &ended
	if runErr != nil {
		w.meta.ExitError = runErr.Error()
	}

	if w.buf != nil {
		_ = w.buf.Flush()
	}
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
		w.buf = nil
	}

	return writeMeta(w.dir, w.meta)
}

func (w *Writer) Dir() string {
	if w == nil {
		return ""
	}
	return w.dir
}

func ListSessions(workDir string) ([]Session, error) {
	root, err := platform.SessionsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	filterDir := ""
	if strings.TrimSpace(workDir) != "" {
		filterDir, err = filepath.Abs(workDir)
		if err != nil {
			return nil, fmt.Errorf("resolve work dir: %w", err)
		}
	}

	sessions := make([]Session, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(root, entry.Name())
		meta, err := LoadMeta(dir)
		if err != nil {
			continue
		}
		if filterDir != "" && !samePath(meta.WorkDir, filterDir) {
			continue
		}
		sessions = append(sessions, Session{Meta: meta, Dir: dir})
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Meta.StartedAt.After(sessions[j].Meta.StartedAt)
	})

	return sessions, nil
}

func LoadMeta(dir string) (Meta, error) {
	path := filepath.Join(dir, metaFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		return Meta{}, err
	}
	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return Meta{}, err
	}
	return meta, nil
}

func ReadLog(dir string) (string, error) {
	path := filepath.Join(dir, logFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func LogPath(dir string) string {
	return filepath.Join(dir, logFilename)
}

func allocateSessionDir(root string) (id, dir string, err error) {
	base := time.Now().Format("20060102-150405")
	for i := 0; i < 100; i++ {
		id = base
		if i > 0 {
			id = fmt.Sprintf("%s-%d", base, i+1)
		}
		dir = filepath.Join(root, id)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return id, dir, nil
		} else if err != nil {
			return "", "", err
		}
	}
	return "", "", fmt.Errorf("could not allocate session id for %s", base)
}

func writeMeta(dir string, meta Meta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, metaFilename)
	return os.WriteFile(path, data, 0o644)
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return aa == bb
}
