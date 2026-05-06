package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"skill-manager/internal/domain"
)

func TestResolveSkillPath(t *testing.T) {
	t.Parallel()

	t.Run("finds skill in global repository", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		skills := NewMockSkillRepository(t)
		projects := NewMockProjectRepository(t)
		projectSkills := NewMockProjectSkillRepository(t)

		skill := domain.Skill{ID: "my-skill", Name: "my-skill", Path: "/global/skills/my-skill"}
		skills.EXPECT().GetByID(ctx, "my-skill").Return(skill, nil)

		path, err := resolveSkillPath(ctx, "my-skill", "", skills, projects, projectSkills)
		require.NoError(t, err)
		require.Equal(t, skill.Path, path)
	})

	t.Run("finds skill in project repository when not global", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		skills := NewMockSkillRepository(t)
		projects := NewMockProjectRepository(t)
		projectSkills := NewMockProjectSkillRepository(t)

		proj := domain.Project{ID: "proj-a", Path: "/workspace/proj-a"}
		skill := domain.Skill{ID: "my-skill", Name: "my-skill", Path: "/workspace/proj-a/skills/my-skill"}

		skills.EXPECT().GetByID(ctx, "my-skill").Return(domain.Skill{}, domain.ErrSkillNotFound)
		projects.EXPECT().GetByID(ctx, "proj-a").Return(proj, nil)
		projectSkills.On("GetByID", ctx, "my-skill", proj).Return(skill, nil).Once()

		path, err := resolveSkillPath(ctx, "my-skill", "proj-a", skills, projects, projectSkills)
		require.NoError(t, err)
		require.Equal(t, skill.Path, path)
	})

	t.Run("returns error when skill not found anywhere", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		skills := NewMockSkillRepository(t)
		projects := NewMockProjectRepository(t)
		projectSkills := NewMockProjectSkillRepository(t)

		proj := domain.Project{ID: "proj-a", Path: "/workspace/proj-a"}

		skills.EXPECT().GetByID(ctx, "missing").Return(domain.Skill{}, domain.ErrSkillNotFound)
		projects.EXPECT().GetByID(ctx, "proj-a").Return(proj, nil)
		projectSkills.On("GetByID", ctx, "missing", proj).Return(domain.Skill{}, domain.ErrSkillNotFound).Once()

		_, err := resolveSkillPath(ctx, "missing", "proj-a", skills, projects, projectSkills)
		require.Error(t, err)
	})
}
