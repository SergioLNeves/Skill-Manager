package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymlinkManager_EnsureLink(t *testing.T) {
	t.Parallel()

	t.Run("creates symlink from source to target", func(t *testing.T) {
		t.Parallel()

		root, source, target := symlinkSetup(t)
		mgr := NewSymlinkManager(root)

		err := mgr.EnsureLink(context.Background(), source, target)

		require.NoError(t, err)
		resolved, err := os.Readlink(target)
		require.NoError(t, err)
		assert.Equal(t, source, resolved)
	})

	t.Run("is idempotent when correct link already exists", func(t *testing.T) {
		t.Parallel()

		root, source, target := symlinkSetup(t)
		mgr := NewSymlinkManager(root)

		require.NoError(t, mgr.EnsureLink(context.Background(), source, target))
		require.NoError(t, mgr.EnsureLink(context.Background(), source, target))

		resolved, err := os.Readlink(target)
		require.NoError(t, err)
		assert.Equal(t, source, resolved)
	})

	t.Run("recreates link when managed link points to wrong source", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		source1 := filepath.Join(root, "skill-a")
		source2 := filepath.Join(root, "skill-b")
		target := filepath.Join(t.TempDir(), "link")

		require.NoError(t, os.MkdirAll(source1, 0o755))
		require.NoError(t, os.MkdirAll(source2, 0o755))

		mgr := NewSymlinkManager(root)
		require.NoError(t, mgr.EnsureLink(context.Background(), source1, target))
		require.NoError(t, mgr.EnsureLink(context.Background(), source2, target))

		resolved, err := os.Readlink(target)
		require.NoError(t, err)
		assert.Equal(t, source2, resolved)
	})

	t.Run("creates parent directories when missing", func(t *testing.T) {
		t.Parallel()

		root, source, _ := symlinkSetup(t)
		mgr := NewSymlinkManager(root)
		target := filepath.Join(t.TempDir(), "deep", "nested", "link")

		err := mgr.EnsureLink(context.Background(), source, target)

		require.NoError(t, err)
		_, statErr := os.Lstat(target)
		assert.NoError(t, statErr)
	})

	t.Run("returns ErrSourceNotFound when source missing", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		mgr := NewSymlinkManager(root)
		source := filepath.Join(root, "missing-skill")
		target := filepath.Join(t.TempDir(), "link")

		err := mgr.EnsureLink(context.Background(), source, target)

		require.ErrorIs(t, err, ErrSourceNotFound)
	})

	t.Run("returns ErrTargetNotManaged when target is a regular file", func(t *testing.T) {
		t.Parallel()

		root, source, _ := symlinkSetup(t)
		mgr := NewSymlinkManager(root)
		target := filepath.Join(t.TempDir(), "regular-file")
		require.NoError(t, os.WriteFile(target, []byte("user content"), 0o644))

		err := mgr.EnsureLink(context.Background(), source, target)

		require.ErrorIs(t, err, ErrTargetNotManaged)
	})

	t.Run("returns ErrTargetNotManaged when target is unmanaged symlink", func(t *testing.T) {
		t.Parallel()

		root, source, _ := symlinkSetup(t)
		mgr := NewSymlinkManager(root)
		outsideDir := t.TempDir()
		target := filepath.Join(t.TempDir(), "link")
		require.NoError(t, os.Symlink(outsideDir, target))

		err := mgr.EnsureLink(context.Background(), source, target)

		require.ErrorIs(t, err, ErrTargetNotManaged)
	})
}

func TestSymlinkManager_RemoveLink(t *testing.T) {
	t.Parallel()

	t.Run("removes managed symlink", func(t *testing.T) {
		t.Parallel()

		root, source, target := symlinkSetup(t)
		mgr := NewSymlinkManager(root)
		require.NoError(t, mgr.EnsureLink(context.Background(), source, target))

		err := mgr.RemoveLink(context.Background(), target)

		require.NoError(t, err)
		_, statErr := os.Lstat(target)
		assert.True(t, os.IsNotExist(statErr))
	})

	t.Run("is no-op when target does not exist", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		mgr := NewSymlinkManager(root)

		err := mgr.RemoveLink(context.Background(), filepath.Join(t.TempDir(), "nonexistent"))
		require.NoError(t, err)
	})

	t.Run("returns ErrTargetNotManaged when target is unmanaged", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		mgr := NewSymlinkManager(root)
		outsideDir := t.TempDir()
		target := filepath.Join(t.TempDir(), "unmanaged-link")
		require.NoError(t, os.Symlink(outsideDir, target))

		err := mgr.RemoveLink(context.Background(), target)

		require.ErrorIs(t, err, ErrTargetNotManaged)
	})
}

func TestSymlinkManager_IsManagedLink(t *testing.T) {
	t.Parallel()

	t.Run("returns true for symlink inside managed root", func(t *testing.T) {
		t.Parallel()

		root, source, target := symlinkSetup(t)
		mgr := NewSymlinkManager(root)
		require.NoError(t, os.Symlink(source, target))

		managed, err := mgr.IsManagedLink(context.Background(), target)

		require.NoError(t, err)
		assert.True(t, managed)
	})

	t.Run("returns false for symlink outside managed root", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		outside := t.TempDir()
		target := filepath.Join(t.TempDir(), "link")
		require.NoError(t, os.Symlink(outside, target))

		mgr := NewSymlinkManager(root)
		managed, err := mgr.IsManagedLink(context.Background(), target)

		require.NoError(t, err)
		assert.False(t, managed)
	})

	t.Run("returns false for regular file", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		target := filepath.Join(t.TempDir(), "regular")
		require.NoError(t, os.WriteFile(target, []byte("data"), 0o644))

		mgr := NewSymlinkManager(root)
		managed, err := mgr.IsManagedLink(context.Background(), target)

		require.NoError(t, err)
		assert.False(t, managed)
	})

	t.Run("handles broken symlink pointing inside managed root", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		// Point to a path inside root that does not exist
		source := filepath.Join(root, "deleted-skill")
		target := filepath.Join(t.TempDir(), "broken-link")
		require.NoError(t, os.Symlink(source, target))

		mgr := NewSymlinkManager(root)
		managed, err := mgr.IsManagedLink(context.Background(), target)

		require.NoError(t, err)
		assert.True(t, managed)
	})
}

// symlinkSetup creates a managed root with one source dir and a target path (not yet created).
func symlinkSetup(t *testing.T) (root, source, target string) {
	t.Helper()
	root = t.TempDir()
	source = filepath.Join(root, "skill-x")
	require.NoError(t, os.MkdirAll(source, 0o755))
	target = filepath.Join(t.TempDir(), "link")
	return root, source, target
}
