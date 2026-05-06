package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	githubfetcher "skill-manager/internal/adapter/github"
	"skill-manager/internal/adapter/managed"
)

// Run dispatches CLI subcommands. args should be os.Args[2:] (after "skills").
func Run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 0
	}
	ctx := context.Background()
	switch args[0] {
	case "add":
		return cmdAdd(ctx, args[1:])
	case "sync":
		return cmdSync(ctx, args[1:])
	case "list":
		return cmdList(args[1:])
	case "remove":
		return cmdRemove(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", args[0])
		printUsage()
		return 1
	}
}

// cmdAdd handles: skills add <owner/repo> [--skill <name>] [--ref <ref>]
func cmdAdd(ctx context.Context, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: skills add <owner/repo> [--skill <name>] [--ref <ref>]")
		return 1
	}
	repo := args[0]
	var skillName, ref string
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--skill":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "--skill requires a value")
				return 1
			}
			i++
			skillName = args[i]
		case "--ref":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "--ref requires a value")
				return 1
			}
			i++
			ref = args[i]
		default:
			if strings.HasPrefix(args[i], "--") {
				fmt.Fprintf(os.Stderr, "unknown flag %q\n", args[i])
				return 1
			}
		}
	}
	if ref == "" {
		ref = "main"
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	fetcher := githubfetcher.NewFetcher()
	repo2 := managed.NewSkillRepository(fetcher)

	fmt.Printf("Fetching %s@%s…\n", repo, ref)
	result, err := repo2.Install(ctx, managed.InstallRequest{
		ProjectDir: cwd,
		Repo:       repo,
		Ref:        ref,
		SkillName:  skillName,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	for _, s := range result.Skills {
		fmt.Printf("  ✓ installed %s  (%s)\n", s.Name, s.Path)
	}
	if len(result.Skills) == 0 {
		fmt.Fprintln(os.Stderr, "no skills found")
		return 1
	}
	return 0
}

// cmdSync reads skills-lock.json and ensures all skills are downloaded.
func cmdSync(ctx context.Context, _ []string) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	lf, err := managed.ReadLockFile(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if len(lf.Skills) == 0 {
		fmt.Println("nothing to sync")
		return 0
	}

	fetcher := githubfetcher.NewFetcher()
	repo := managed.NewSkillRepository(fetcher)

	for name, entry := range lf.Skills {
		ref := entry.Ref
		if ref == "" {
			ref = "main"
		}
		fmt.Printf("  syncing %s from %s@%s…\n", name, entry.Source, ref)
		_, err := repo.Install(ctx, managed.InstallRequest{
			ProjectDir: cwd,
			Repo:       entry.Source,
			Ref:        ref,
			SkillName:  name,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "  error syncing %s: %v\n", name, err)
		}
	}
	fmt.Println("sync complete")
	return 0
}

// cmdList prints the skills in the lock file.
func cmdList(_ []string) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	lf, err := managed.ReadLockFile(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if len(lf.Skills) == 0 {
		fmt.Println("no skills installed")
		return 0
	}
	fmt.Printf("%-30s %-40s %s\n", "NAME", "REPO", "REF")
	for name, e := range lf.Skills {
		ref := e.Ref
		if len(ref) > 8 {
			ref = ref[:8]
		}
		fmt.Printf("%-30s %-40s %s\n", name, e.Source, ref)
	}
	return 0
}

// cmdRemove removes a skill from the lock file.
func cmdRemove(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: skills remove <name>")
		return 1
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	lf, err := managed.ReadLockFile(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	name := args[0]
	if _, ok := lf.Skills[name]; !ok {
		fmt.Fprintf(os.Stderr, "skill %q not found in lock file\n", name)
		return 1
	}
	if err := lf.Remove(name); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("removed %s\n", name)
	return 0
}

func printUsage() {
	fmt.Println(`Usage: skill-manager skills <command> [options]

Commands:
  add <owner/repo> [--skill <name>] [--ref <ref>]   Install skill(s) from GitHub
  sync                                                Re-download skills from lock file
  list                                                List installed skills
  remove <name>                                       Remove skill from lock file`)
}
