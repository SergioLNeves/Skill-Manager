package filesystem

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrTargetNotManaged is returned when a target path exists but is not a
// symlink managed by skills-manager (i.e. not pointing inside managedRoot).
var ErrTargetNotManaged = errors.New("symlink manager: target exists and is not managed")

// ErrSourceNotFound is returned when the source path does not exist.
var ErrSourceNotFound = errors.New("symlink manager: source path not found")

// SymlinkManager creates and removes filesystem symlinks for agent adapters.
// Operations are limited to Linux and macOS; Windows is not supported.
type SymlinkManager struct {
	managedRoot string // path that managed symlinks must point inside of
}

// NewSymlinkManager creates a manager that considers symlinks managed when
// their resolved target is inside managedRoot.
func NewSymlinkManager(managedRoot string) *SymlinkManager {
	return &SymlinkManager{managedRoot: managedRoot}
}

// EnsureLink creates a symlink at target pointing to source.
// It is idempotent: no-op when the correct link already exists.
// Returns ErrTargetNotManaged if something unmanaged already exists at target.
// Returns ErrSourceNotFound if source does not exist.
func (m *SymlinkManager) EnsureLink(_ context.Context, source, target string) error {
	if _, err := os.Lstat(source); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrSourceNotFound, source)
		}
		return fmt.Errorf("symlink manager: stat source %s: %w", source, err)
	}

	info, err := os.Lstat(target)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("symlink manager: stat target %s: %w", target, err)
	}

	if err == nil {
		// target exists
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("%w: %s is a regular file or directory", ErrTargetNotManaged, target)
		}

		managed, err := m.IsManagedLink(context.Background(), target)
		if err != nil {
			return fmt.Errorf("symlink manager: check managed %s: %w", target, err)
		}
		if !managed {
			return fmt.Errorf("%w: %s points outside managed root", ErrTargetNotManaged, target)
		}

		resolved, _ := os.Readlink(target)
		if resolved == source {
			return nil // already correct
		}
		// managed but wrong destination — recreate
		if err := os.Remove(target); err != nil {
			return fmt.Errorf("symlink manager: remove stale link %s: %w", target, err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("symlink manager: mkdir %s: %w", filepath.Dir(target), err)
	}

	if err := os.Symlink(source, target); err != nil {
		return fmt.Errorf("symlink manager: create link %s -> %s: %w", target, source, err)
	}
	return nil
}

// RemoveLink removes the symlink at target only when IsManagedLink is true.
func (m *SymlinkManager) RemoveLink(_ context.Context, target string) error {
	managed, err := m.IsManagedLink(context.Background(), target)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // already gone
		}
		return fmt.Errorf("symlink manager: check managed %s: %w", target, err)
	}
	if !managed {
		return fmt.Errorf("%w: refusing to remove unmanaged path %s", ErrTargetNotManaged, target)
	}
	if err := os.Remove(target); err != nil {
		return fmt.Errorf("symlink manager: remove %s: %w", target, err)
	}
	return nil
}

// IsManagedLink reports whether target is a symlink whose resolved destination
// is inside the managed root.
func (m *SymlinkManager) IsManagedLink(_ context.Context, target string) (bool, error) {
	info, err := os.Lstat(target)
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false, nil
	}

	dest, err := filepath.EvalSymlinks(target)
	if err != nil {
		// broken symlink — resolve without following
		raw, readErr := os.Readlink(target)
		if readErr != nil {
			return false, fmt.Errorf("symlink manager: readlink %s: %w", target, readErr)
		}
		dest = raw
	}

	abs, err := filepath.Abs(dest)
	if err != nil {
		return false, fmt.Errorf("symlink manager: abs path %s: %w", dest, err)
	}

	root := filepath.Clean(m.managedRoot)
	return strings.HasPrefix(abs, root+string(filepath.Separator)) || abs == root, nil
}
