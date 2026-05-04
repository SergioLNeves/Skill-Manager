package usecase

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopySkillRequest carries the parameters for a skill copy operation.
type CopySkillRequest struct {
	SkillID         string
	SourceProjectID string // empty when skill is global
	TargetProjectID string
}

// CopySkill physically copies a skill directory into a target project's skills/ folder.
type CopySkill struct {
	skills        SkillRepository
	projects      ProjectRepository
	projectSkills ProjectSkillRepository
}

func NewCopySkill(skills SkillRepository, projects ProjectRepository, projectSkills ProjectSkillRepository) *CopySkill {
	return &CopySkill{skills: skills, projects: projects, projectSkills: projectSkills}
}

func (uc *CopySkill) Execute(ctx context.Context, req CopySkillRequest) error {
	srcPath, err := resolveSkillPath(ctx, req.SkillID, req.SourceProjectID, uc.skills, uc.projects, uc.projectSkills)
	if err != nil {
		return fmt.Errorf("copy skill: resolve source: %w", err)
	}

	target, err := uc.projects.GetByID(ctx, req.TargetProjectID)
	if err != nil {
		return fmt.Errorf("copy skill: resolve target project: %w", err)
	}

	srcDir := filepath.Dir(srcPath)
	dstDir := filepath.Join(target.Path, "skills", filepath.Base(srcDir))

	if _, err := os.Stat(dstDir); err == nil {
		return fmt.Errorf("copy skill: destination already exists: %s", dstDir)
	}

	return copyDir(srcDir, dstDir)
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("copyDir mkdir %s: %w", dst, err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("copyDir readdir %s: %w", src, err)
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	var closeErr error
	defer func() {
		closeErr = out.Close()
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return closeErr
}

// resolveSkillPath returns the absolute path to SKILL.md for the given skillID,
// searching first global then project repositories.
func resolveSkillPath(ctx context.Context, skillID, projectID string, skills SkillRepository, projects ProjectRepository, projectSkills ProjectSkillRepository) (string, error) {
	if s, err := skills.GetByID(ctx, skillID); err == nil {
		return s.Path, nil
	}
	if projectID == "" {
		return "", fmt.Errorf("skill %s not found", skillID)
	}
	project, err := projects.GetByID(ctx, projectID)
	if err != nil {
		return "", err
	}
	s, err := projectSkills.GetByID(ctx, skillID, project)
	if err != nil {
		return "", err
	}
	return s.Path, nil
}
