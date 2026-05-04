package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"skill-manager/internal/domain"
	"skill-manager/internal/usecase"
)

// openTestDB returns a fresh in-memory database for each test.
func openTestDB(t *testing.T) *testDB {
	t.Helper()
	db, err := OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return &testDB{
		projects:    NewProjectRepository(db),
		activations: NewActivationRepository(db),
	}
}

type testDB struct {
	projects    *ProjectRepository
	activations *ActivationRepository
}

// --- ProjectRepository ---

func TestProjectRepository_SaveAndGet(t *testing.T) {
	t.Parallel()
	tdb := openTestDB(t)

	p := domain.Project{
		ID:             "proj-1",
		Name:           "my-service",
		Path:           "/home/user/dev/my-service",
		DetectedAgents: []domain.Agent{domain.AgentClaude, domain.AgentCopilot},
		AddedAt:        time.Now().UTC().Truncate(time.Second),
	}

	require.NoError(t, tdb.projects.Save(context.Background(), p))

	got, err := tdb.projects.GetByID(context.Background(), "proj-1")
	require.NoError(t, err)
	assert.Equal(t, p.ID, got.ID)
	assert.Equal(t, p.Name, got.Name)
	assert.Equal(t, p.Path, got.Path)
	assert.ElementsMatch(t, p.DetectedAgents, got.DetectedAgents)
}

func TestProjectRepository_List(t *testing.T) {
	t.Parallel()
	tdb := openTestDB(t)

	for _, p := range []domain.Project{
		{ID: "a", Name: "alpha", Path: "/alpha", AddedAt: time.Now().UTC()},
		{ID: "b", Name: "beta", Path: "/beta", AddedAt: time.Now().UTC()},
	} {
		require.NoError(t, tdb.projects.Save(context.Background(), p))
	}

	list, err := tdb.projects.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestProjectRepository_Save_Upsert(t *testing.T) {
	t.Parallel()
	tdb := openTestDB(t)

	p := domain.Project{ID: "p1", Name: "old", Path: "/p1", AddedAt: time.Now().UTC()}
	require.NoError(t, tdb.projects.Save(context.Background(), p))

	p.Name = "new"
	p.DetectedAgents = []domain.Agent{domain.AgentClaude}
	require.NoError(t, tdb.projects.Save(context.Background(), p))

	got, err := tdb.projects.GetByID(context.Background(), "p1")
	require.NoError(t, err)
	assert.Equal(t, "new", got.Name)
	assert.Equal(t, []domain.Agent{domain.AgentClaude}, got.DetectedAgents)
}

func TestProjectRepository_Delete(t *testing.T) {
	t.Parallel()
	tdb := openTestDB(t)

	p := domain.Project{ID: "del", Name: "del", Path: "/del", AddedAt: time.Now().UTC()}
	require.NoError(t, tdb.projects.Save(context.Background(), p))
	require.NoError(t, tdb.projects.Delete(context.Background(), "del"))

	_, err := tdb.projects.GetByID(context.Background(), "del")
	require.ErrorIs(t, err, domain.ErrProjectNotFound)
}

func TestProjectRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	tdb := openTestDB(t)

	_, err := tdb.projects.GetByID(context.Background(), "nonexistent")
	require.ErrorIs(t, err, domain.ErrProjectNotFound)
}

func TestProjectRepository_Delete_NotFound(t *testing.T) {
	t.Parallel()
	tdb := openTestDB(t)

	err := tdb.projects.Delete(context.Background(), "ghost")
	require.ErrorIs(t, err, domain.ErrProjectNotFound)
}

// --- ActivationRepository ---

func TestActivationRepository_SaveAndList(t *testing.T) {
	t.Parallel()
	tdb := openTestDB(t)
	saveProject(t, tdb, "proj-a", "/proj-a")

	a := domain.Activation{
		SkillID:   "skill-1",
		Agent:     domain.AgentClaude,
		Scope:     domain.ScopeGlobal,
		AppliedAt: time.Now().UTC().Truncate(time.Second),
	}

	saved, err := tdb.activations.Save(context.Background(), a)
	require.NoError(t, err)
	assert.NotZero(t, saved.ID)

	list, err := tdb.activations.List(context.Background(), usecase.ActivationFilter{})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, saved.ID, list[0].ID)
	assert.Equal(t, "skill-1", list[0].SkillID)
	assert.Equal(t, domain.AgentClaude, list[0].Agent)
}

func TestActivationRepository_List_Filter(t *testing.T) {
	t.Parallel()
	tdb := openTestDB(t)
	saveProject(t, tdb, "p1", "/p1")
	pid := "p1"

	activations := []domain.Activation{
		{SkillID: "s1", Agent: domain.AgentClaude, Scope: domain.ScopeGlobal, AppliedAt: time.Now().UTC()},
		{SkillID: "s1", Agent: domain.AgentCopilot, Scope: domain.ScopeProject, ProjectID: &pid, AppliedAt: time.Now().UTC()},
		{SkillID: "s2", Agent: domain.AgentClaude, Scope: domain.ScopeGlobal, AppliedAt: time.Now().UTC()},
	}
	for _, a := range activations {
		_, err := tdb.activations.Save(context.Background(), a)
		require.NoError(t, err)
	}

	t.Run("filter by skill", func(t *testing.T) {
		list, err := tdb.activations.List(context.Background(), usecase.ActivationFilter{SkillID: "s1"})
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("filter by agent", func(t *testing.T) {
		list, err := tdb.activations.List(context.Background(), usecase.ActivationFilter{Agent: domain.AgentClaude})
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("filter by scope", func(t *testing.T) {
		list, err := tdb.activations.List(context.Background(), usecase.ActivationFilter{Scope: domain.ScopeGlobal})
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})
}

func TestActivationRepository_Delete(t *testing.T) {
	t.Parallel()
	tdb := openTestDB(t)

	saved, err := tdb.activations.Save(context.Background(), domain.Activation{
		SkillID: "s1", Agent: domain.AgentClaude, Scope: domain.ScopeGlobal, AppliedAt: time.Now().UTC(),
	})
	require.NoError(t, err)

	require.NoError(t, tdb.activations.Delete(context.Background(), saved.ID))

	list, err := tdb.activations.List(context.Background(), usecase.ActivationFilter{})
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestActivationRepository_FindConflict(t *testing.T) {
	t.Parallel()
	tdb := openTestDB(t)
	saveProject(t, tdb, "proj-1", "/proj-1")
	pid := "proj-1"

	global := domain.Activation{
		SkillID: "go-review", Agent: domain.AgentClaude, Scope: domain.ScopeGlobal, AppliedAt: time.Now().UTC(),
	}
	project := domain.Activation{
		SkillID: "go-review", Agent: domain.AgentClaude, Scope: domain.ScopeProject, ProjectID: &pid, AppliedAt: time.Now().UTC(),
	}

	t.Run("returns nil when no conflict", func(t *testing.T) {
		conflict, err := tdb.activations.FindConflict(context.Background(), "go-review", domain.AgentClaude, "proj-1")
		require.NoError(t, err)
		assert.Nil(t, conflict)
	})

	_, err := tdb.activations.Save(context.Background(), global)
	require.NoError(t, err)

	t.Run("returns nil when only global exists", func(t *testing.T) {
		conflict, err := tdb.activations.FindConflict(context.Background(), "go-review", domain.AgentClaude, "proj-1")
		require.NoError(t, err)
		assert.Nil(t, conflict)
	})

	_, err = tdb.activations.Save(context.Background(), project)
	require.NoError(t, err)

	t.Run("returns conflict when both global and project exist", func(t *testing.T) {
		conflict, err := tdb.activations.FindConflict(context.Background(), "go-review", domain.AgentClaude, "proj-1")
		require.NoError(t, err)
		require.NotNil(t, conflict)
		assert.Equal(t, "go-review", conflict.SkillID)
		assert.Equal(t, domain.AgentClaude, conflict.Agent)
		assert.NotNil(t, conflict.GlobalActivation)
		assert.NotNil(t, conflict.ProjectActivation)
	})
}

func TestActivationRepository_DeleteProject_CascadesActivations(t *testing.T) {
	t.Parallel()
	tdb := openTestDB(t)
	saveProject(t, tdb, "proj-x", "/proj-x")
	pid := "proj-x"

	_, err := tdb.activations.Save(context.Background(), domain.Activation{
		SkillID: "s1", Agent: domain.AgentClaude, Scope: domain.ScopeProject,
		ProjectID: &pid, AppliedAt: time.Now().UTC(),
	})
	require.NoError(t, err)

	require.NoError(t, tdb.projects.Delete(context.Background(), "proj-x"))

	list, err := tdb.activations.List(context.Background(), usecase.ActivationFilter{})
	require.NoError(t, err)
	assert.Empty(t, list)
}

func saveProject(t *testing.T, tdb *testDB, id, path string) {
	t.Helper()
	require.NoError(t, tdb.projects.Save(context.Background(), domain.Project{
		ID: id, Name: id, Path: path, AddedAt: time.Now().UTC(),
	}))
}
