package watcher

import (
	"context"
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"
)

// SkillsWatcher monitors the central skills directory and calls onChange
// whenever a skill is added, removed, or renamed.
type SkillsWatcher struct {
	dir      string
	onChange func()
	watcher  *fsnotify.Watcher
}

// NewSkillsWatcher creates a watcher for skillsDir.
func NewSkillsWatcher(skillsDir string, onChange func()) (*SkillsWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err = w.Add(skillsDir); err != nil {
		w.Close()
		return nil, err
	}
	return &SkillsWatcher{dir: skillsDir, onChange: onChange, watcher: w}, nil
}

// Run blocks until ctx is cancelled, debouncing filesystem events.
func (sw *SkillsWatcher) Run(ctx context.Context) {
	defer sw.watcher.Close()

	var debounce <-chan time.Time

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-sw.watcher.Events:
			if !ok {
				return
			}
			slog.Debug("skills dir event", "op", event.Op.String(), "name", event.Name)
			debounce = time.After(300 * time.Millisecond)
		case err, ok := <-sw.watcher.Errors:
			if !ok {
				return
			}
			slog.Warn("skills watcher error", "err", err)
		case <-debounce:
			slog.Info("skills directory changed, refreshing")
			sw.onChange()
			debounce = nil
		}
	}
}
