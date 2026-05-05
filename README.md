# skill-manager

A desktop application built with Wails v2 (Go backend + React/TypeScript frontend) for managing skills and activating them in local development projects for AI agents.

A skill is any directory containing a `SKILL.md` file with a YAML frontmatter header and a markdown body. skill-manager discovers these directories, lets you organize them, and copies them into the right place for Claude or GitHub Copilot to pick up.

## Download and install (pre-built binaries)

Download the latest release from the [Releases page](../../releases/latest):

| Platform | File |
|---|---|
| Linux x86-64 | `skill-manager-linux-amd64.tar.gz` |
| macOS Intel | `skill-manager-darwin-amd64.zip` |
| macOS Apple Silicon | `skill-manager-darwin-arm64.zip` |

### Linux

```bash
tar -xzf skill-manager-linux-amd64.tar.gz
chmod +x skill-manager
./skill-manager
```

**Prerequisite**: WebKit2GTK and GTK3 must be installed.

```bash
# Debian / Ubuntu
sudo apt install libwebkit2gtk-4.1-0 libgtk-3-0

# Fedora
sudo dnf install webkit2gtk4.1 gtk3

# Arch Linux
sudo pacman -S webkit2gtk gtk3
```

### macOS

1. Extract the `.zip` and move `skill-manager.app` to `/Applications`.
2. The app is **not signed or notarized**. On first launch macOS will block it. Remove the quarantine flag:

```bash
xattr -dr com.apple.quarantine /Applications/skill-manager.app
```

Or go to **System Settings → Privacy & Security** and click **Open Anyway** after the first blocked attempt.

---

## How it works

Skills are discovered by scanning configured directories recursively. Any folder containing a `SKILL.md` is treated as a skill.

There are two kinds of skills:

- **Global skills**: discovered from the global skill roots you configure in Settings. Available to any project.
- **Project skills**: live inside a project's own `skills/` or `.claude/skills/` directory. Visible only to that project.

Activating a skill for a project copies the skill directory into `.claude/skills/` (for Claude) or `.github/skills/` (for GitHub Copilot) inside the project folder. The activation is also recorded in a local SQLite database.

Skills can be grouped by category using the `category:` field in the `SKILL.md` frontmatter. The Categories page shows all skills grouped by category and lets you copy an entire category into a project at once.

The Doctor page checks for inconsistencies between the database and the filesystem — orphaned activations, missing projects, broken symlinks, directories that no longer exist — and offers automatic fixes where possible.

## Architecture

The backend is written in Go using a layered structure: domain types, use cases, adapters (filesystem, SQLite), and bindings. A small DI container in `internal/di/container.go` wires everything together. Public methods on the `App` struct in `app.go` are automatically exposed to the frontend by Wails.

The frontend uses React with TypeScript, TanStack Router (file-based routes under `frontend/src/routes/`), TanStack Query for server state, and Tailwind CSS with shadcn-style components.

Storage is SQLite via `modernc.org/sqlite`. Migrations are embedded in the binary and run automatically on startup from `internal/adapter/persistence/migrations/`.

## Requirements (build from source)

- Go 1.22 or later
- Node.js 20 or later and npm
- Wails CLI v2: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Platform-specific build dependencies listed at https://wails.io/docs/gettingstarted/installation

## Install and run

```bash
git clone <repo-url>
cd skill-manager
cd frontend && npm install && cd ..
wails dev      # development mode with hot reload
```

To build a standalone binary:

```bash
wails build
# output: build/bin/skill-manager
```

## Usage

1. Open **Settings** and add at least one Workspace Root — a folder containing your development projects. Optionally add Global Skill Roots where your shared skills live.
2. Open **Projects**, click Scan to discover projects under the configured roots, and add the ones you want to manage.
3. Open **Skills** to browse all discovered skills. Each skill card shows its name, description, and category. Click a skill to see its full content and the locations where it exists.
4. Open **Categories** to browse skills grouped by the `category:` frontmatter field. Use "Add to project" on a category to copy all its skills into a project at once.
5. Open a project from the Projects page and toggle individual skills on or off per agent (Claude or Copilot). Use **Reset Skills** to remove all copied skills and clear all activations for that project.
6. Run **Doctor** periodically to detect and fix drift between the database and the filesystem.

## Skill format

A skill is a directory containing a `SKILL.md` file:

```
my-skill/
  SKILL.md
  other-files...
```

The `SKILL.md` file uses YAML frontmatter followed by a markdown body:

```markdown
---
name: My Skill
description: A short description of what this skill does
category: Backend
---

# Body

Instructions, context, or rules that the agent should follow when this skill is active.
```

The `name`, `description`, and `category` fields are optional but recommended. If `name` is omitted, the directory name is used.

## Development notes

After adding or changing public methods in `app.go`, regenerate the TypeScript bindings:

```bash
wails generate module
```

This updates `frontend/wailsjs/go/main/App.d.ts` and `frontend/wailsjs/go/main/App.js`.

The frontend has no automated tests. The backend has unit tests using testify with mocks generated by mockery. Run backend tests with:

```bash
go test ./...
```
