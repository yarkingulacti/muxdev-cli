package logs

import "os"

// RemoveSessions deletes all saved runtime session directories for workDir.
func RemoveSessions(workDir string) (int, error) {
	sessions, err := ListSessions(workDir)
	if err != nil {
		return 0, err
	}

	removed := 0
	for _, session := range sessions {
		if err := os.RemoveAll(session.Dir); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}
