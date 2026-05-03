# Skills Manager — Implementation Plan

> Manager offline de Skills (Globais e por Projeto) para agentes de IA, com GUI Wails + React/TS.

---

## Sumário

- [Visão geral](#visão-geral)
- [Decisões arquiteturais](#decisões-arquiteturais)
- [Stack & convenções](#stack--convenções)
- [Estrutura de pastas](#estrutura-de-pastas)
- [Modelo de domínio](#modelo-de-domínio)
- [Fases de implementação](#fases-de-implementação)
  - [Fase 0 — Bootstrap](#fase-0--bootstrap)
  - [Fase 1 — Domain core](#fase-1--domain-core)
  - [Fase 2 — Filesystem adapters](#fase-2--filesystem-adapters)
  - [Fase 3 — Agent adapters (Claude + Copilot)](#fase-3--agent-adapters-claude--copilot)
  - [Fase 4 — Persistência (SQLite)](#fase-4--persistência-sqlite)
  - [Fase 5 — Wails bindings](#fase-5--wails-bindings)
  - [Fase 6 — Frontend (React/TS + styled-components)](#fase-6--frontend-react--ts--styled-components)
  - [Fase 7 — Polimento & doctor](#fase-7--polimento--doctor)
- [Riscos & armadilhas](#riscos--armadilhas)
- [Pós-MVP](#pós-mvp)

---

## Visão geral

### Problema

Conforme as skills crescem em número, fica difícil saber:
- Quais estão ativas globalmente.
- Quais estão ativas em um projeto específico.
- Se há conflito entre global × projeto.
- Quais agentes (Claude, Copilot, etc.) estão consumindo o quê.

### Solução

Um app desktop offline (Wails) que:
1. Mantém um **repositório central de skills** em `~/.skills-manager/skills/`.
2. Mantém um **registry SQLite** com projetos cadastrados e estado de ativação.
3. Aplica o estado desejado no filesystem via **agent adapters** (symlink pro Claude, arquivo concatenado pro Copilot).
4. Detecta conflitos e força resolução explícita pelo usuário.
5. É 100% transparente pros agentes — eles continuam lendo seus paths nativos.

### Fluxo de uso típico

1. Usuário abre o app.
2. Vê lista de skills disponíveis (lidas do repositório central).
3. Vê lista de projetos cadastrados (sugeridos via scan + adicionados manualmente).
4. Marca skill X como ativa globalmente pro Claude → manager cria symlink em `~/.claude/skills/x/`.
5. Marca skill Y como ativa no projeto Z pro Copilot → manager regenera `<Z>/.github/copilot-instructions.md`.
6. Tenta marcar skill X também no projeto Z pro Claude → manager detecta conflito (já está global) e abre modal.

---

## Decisões arquiteturais

| Decisão | Escolha | Motivo |
|---|---|---|
| Plataformas suportadas | **Linux + macOS apenas** (Windows fora de escopo) | Symlinks são primeira classe em Unix; Windows traz complexidade de Developer Mode/privilégios que não vale pro MVP |
| Modelo de ativação | Symlinks (operação direta, sem fallback) | Transparente pro agente, sem duplicação |
| Conflito global × projeto | Erro explícito, usuário resolve | Evita comportamento mágico/surpreendente |
| Detecção de projetos | Híbrido (scan sugere, usuário confirma + adiciona manual) | Equilíbrio entre conveniência e controle |
| UI | GUI Wails apenas (sem CLI no MVP) | Foco no MVP |
| Storage | SQLite | Queries de conflito ficam triviais; migrations claras |
| Agentes MVP | Claude + Copilot | Valida abstração com dois modelos opostos (pasta vs arquivo único) |
| Arquitetura backend | Hexagonal (domain / ports / adapters) | Padrão familiar do Serjol; testabilidade |
| Frontend | React + TS + styled-components + Container Pattern | Padrão atual do Serjol |

---

## Stack & convenções

### Plataformas suportadas

- **Linux** (testado em Ubuntu 22.04+, Fedora 40+, Arch).
- **macOS** (testado em macOS 13+, Apple Silicon e Intel).
- **Windows: não suportado no MVP.** Build não é distribuído. Pós-MVP avaliar suporte com fallback de cópia.

Justificativa: symlinks em Unix são operação trivial e sem privilégio; em Windows exigem Developer Mode ou admin. Tirar Windows do MVP simplifica o `SymlinkManager`, elimina necessidade de capability check no startup e permite que `os.Symlink` seja chamado direto sem branching por SO.

### Backend (Go)

- **Wails v2** — runtime desktop.
- **Go 1.22+**.
- **`samber/do`** — DI (consistente com seus outros projetos Go).
- **`mattn/go-sqlite3`** ou **`modernc.org/sqlite`** (puro Go, sem cgo — mais fácil pra distribuir binário). **Decisão: `modernc.org/sqlite`** pra evitar dor de cabeça com cgo no build do Wails.
- **`testify`** + **`mockery`** — testes (padrão Serjol).
- **`spf13/afero`** — abstração de filesystem (facilita testes de symlink sem tocar disco real).

### Frontend (TypeScript / React)

- **React 18** + **TypeScript estrito**.
- **styled-components** com `ThemeProvider`, convenção `Styled*`, helper `css`, `attrs` quando aplicável.
- **TanStack Query** (mesmo offline, pra cache + invalidação dos bindings Wails — eles são async como qualquer fetch).
- **Container Pattern** (5 camadas: Types → Infra → Pages → Routes → Components) com ESLint boundaries.
- **shadcn/ui + Tailwind 4** opcional pra primitivos (modal, dialog) — mas estilo principal via styled-components.

### Convenções

- **Commits**: Conventional Commits 1.0.0.
- **Go testing**: same package, `t.Run`, `t.Parallel()`, AAA, mockery + testify.
- **Estrutura React**: `components/styled/`, `styles/theme.ts`, `pages/`, `utils/`, hooks em `hooks/useQueries/` e `hooks/useMutations/`.

---

## Estrutura de pastas

```
skills-manager/
├── IMPLEMENTATION_PLAN.md
├── README.md
├── wails.json
├── go.mod
├── main.go                          # entrypoint Wails
├── app.go                            # bridge Wails (chama os bindings)
│
├── internal/
│   ├── domain/                       # entidades e regras puras (sem I/O)
│   │   ├── skill.go
│   │   ├── project.go
│   │   ├── activation.go
│   │   ├── conflict.go
│   │   ├── agent.go                  # tipo Agent (Claude, Copilot, ...)
│   │   └── errors.go
│   │
│   ├── usecase/                      # casos de uso (orquestração)
│   │   ├── list_skills.go
│   │   ├── register_project.go
│   │   ├── scan_projects.go
│   │   ├── activate_skill.go
│   │   ├── deactivate_skill.go
│   │   ├── resolve_conflict.go
│   │   ├── doctor.go                 # valida symlinks órfãos
│   │   └── ports.go                  # interfaces (SkillRepo, ProjectRepo, AgentAdapter, ...)
│   │
│   ├── adapter/
│   │   ├── persistence/              # SQLite
│   │   │   ├── sqlite.go
│   │   │   ├── migrations/
│   │   │   ├── project_repo.go
│   │   │   ├── activation_repo.go
│   │   │   └── *_test.go
│   │   │
│   │   ├── filesystem/               # leitura/escrita de skills
│   │   │   ├── skill_repository.go   # lê ~/.skills-manager/skills/
│   │   │   ├── project_scanner.go    # scan híbrido
│   │   │   ├── symlink_manager.go    # cria/remove symlinks (com fallback)
│   │   │   └── *_test.go
│   │   │
│   │   └── agent/                    # adapters por agente
│   │       ├── claude.go             # ~/.claude/skills/ + <proj>/.claude/skills/
│   │       ├── copilot.go            # <proj>/.github/copilot-instructions.md
│   │       └── *_test.go
│   │
│   ├── di/
│   │   └── container.go              # wiring com samber/do
│   │
│   └── binding/                      # Wails bindings (interface app ↔ frontend)
│       ├── skills_binding.go
│       ├── projects_binding.go
│       └── activation_binding.go
│
├── frontend/
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── src/
│       ├── types/                    # camada Types
│       │   ├── skill.ts
│       │   ├── project.ts
│       │   └── activation.ts
│       │
│       ├── infra/                    # camada Infra (binding calls + queries)
│       │   ├── bindings.ts           # wrappers tipados sobre window.go.*
│       │   └── queries/
│       │       ├── useSkillsQuery.ts
│       │       ├── useProjectsQuery.ts
│       │       └── useActivationMutation.ts
│       │
│       ├── pages/                    # camada Pages (orquestração)
│       │   ├── SkillsPage/
│       │   ├── ProjectsPage/
│       │   ├── ProjectDetailPage/
│       │   └── DoctorPage/
│       │
│       ├── routes/                   # camada Routes (TanStack Router)
│       │   └── ...
│       │
│       ├── components/               # camada Components
│       │   ├── styled/
│       │   ├── SkillCard/
│       │   ├── ProjectCard/
│       │   ├── ActivationToggle/
│       │   └── ConflictModal/
│       │
│       └── styles/
│           ├── theme.ts
│           └── globalStyles.ts
│
└── build/                            # output Wails
```

---

## Modelo de domínio

### Entidades

```go
// domain/skill.go
type Skill struct {
    ID          string    // hash estável do path relativo
    Name        string    // diretório base (ex: "pt-review")
    Path        string    // path absoluto no repositório central
    Description string    // extraído do YAML frontmatter do SKILL.md
    UpdatedAt   time.Time
}

// domain/agent.go
type Agent string

const (
    AgentClaude  Agent = "claude"
    AgentCopilot Agent = "copilot"
)

// domain/project.go
type Project struct {
    ID              string
    Name            string
    Path            string    // path absoluto
    DetectedAgents  []Agent   // detectados via scan ou marcados manualmente
    AddedAt         time.Time
}

// domain/activation.go
type Scope string

const (
    ScopeGlobal  Scope = "global"
    ScopeProject Scope = "project"
)

type Activation struct {
    ID        int64
    SkillID   string
    Agent     Agent
    Scope     Scope
    ProjectID *string   // nil quando Scope == Global
    AppliedAt time.Time
}

// domain/conflict.go
type Conflict struct {
    SkillID    string
    Agent      Agent
    GlobalActivation  *Activation
    ProjectActivation *Activation
}
```

### Invariantes

1. Não pode existir mais de uma `Activation` com a mesma tupla `(skill_id, agent, scope, project_id)`.
2. Quando `Scope == ScopeGlobal`, `ProjectID` é obrigatoriamente `nil`.
3. Quando `Scope == ScopeProject`, `ProjectID` é obrigatoriamente preenchido e o projeto deve existir.
4. Ativação só é válida se o agente correspondente foi detectado/declarado no projeto (pra escopo project).

### Regra de conflito

> Existe conflito quando, pra mesma `(skill_id, agent)`, há ativação global **e** ativação no projeto que está sendo afetado.

A resolução é uma escolha explícita: **manter global**, **promover projeto a override** (mantém ambos no DB mas marca o de projeto como override e o global é "shadowed" naquele projeto), ou **cancelar a operação**.

---

## Fases de implementação

### Fase 0 — Bootstrap

**Objetivo**: ambiente rodando, hello-world Wails + React/TS + styled-components.

**Tarefas**:
- `wails init -n skills-manager -t react-ts`.
- Adicionar styled-components + ThemeProvider básico.
- Configurar ESLint com regras de boundaries (Container Pattern).
- Configurar Prettier.
- Setup Go modules: `samber/do`, `testify`, `modernc.org/sqlite`, `spf13/afero`, `mockery`.
- Estrutura de pastas conforme acima (vazia, com `.gitkeep`).
- Makefile / Taskfile com targets: `dev`, `build`, `test`, `lint`, `mocks`.
- README com instruções de dev.
- CI mínimo (GitHub Actions): `go test`, `go vet`, `pnpm lint`, `pnpm tsc --noEmit` rodando em matriz `[ubuntu-latest, macos-latest]`. Windows explicitamente excluído.

**Critério de pronto**: `wails dev` abre janela com tela "Hello Skills Manager" estilizada via styled-components.

**Commit**: `chore: bootstrap wails react-ts project`

---

### Fase 1 — Domain core

**Objetivo**: modelar entidades + casos de uso puros, com testes 100% offline (sem I/O).

**Tarefas**:
- Implementar entidades em `internal/domain/`.
- Implementar erros tipados em `domain/errors.go` (`ErrConflict`, `ErrSkillNotFound`, `ErrProjectNotFound`, `ErrInvalidScope`, etc.).
- Definir **portas** em `internal/usecase/ports.go`:
  ```go
  type SkillRepository interface {
      List(ctx context.Context) ([]Skill, error)
      GetByID(ctx context.Context, id string) (Skill, error)
  }

  type ProjectRepository interface {
      List(ctx context.Context) ([]Project, error)
      GetByID(ctx context.Context, id string) (Project, error)
      Save(ctx context.Context, p Project) error
      Delete(ctx context.Context, id string) error
  }

  type ActivationRepository interface {
      List(ctx context.Context, filter ActivationFilter) ([]Activation, error)
      Save(ctx context.Context, a Activation) error
      Delete(ctx context.Context, id int64) error
      FindConflict(ctx context.Context, skillID string, agent Agent, projectID string) (*Conflict, error)
  }

  type AgentAdapter interface {
      Agent() Agent
      ApplyGlobal(ctx context.Context, activeSkills []Skill) error
      ApplyProject(ctx context.Context, project Project, activeSkills []Skill) error
      CapabilityCheck(ctx context.Context) error  // verifica permissão de escrita no path do agente
  }

  type ProjectScanner interface {
      Scan(ctx context.Context, roots []string) ([]ProjectCandidate, error)
  }
  ```
- Implementar casos de uso (orquestração apenas, **sem I/O direto** — só chamando portas):
  - `ListSkills`
  - `RegisterProject` (manual)
  - `ScanProjects` (sugestão via ProjectScanner)
  - `ActivateSkill` (faz check de conflito, retorna `Conflict` se houver)
  - `DeactivateSkill`
  - `ResolveConflict`
  - `Doctor` (valida estado consistente)
- Mockery em todas as portas.
- Testes de cada caso de uso cobrindo happy + erro + conflito.

**Critério de pronto**: `go test ./internal/domain/... ./internal/usecase/...` verde com >85% cobertura.

**Commits**:
- `feat: add domain entities for skills and projects`
- `feat: add usecase ports`
- `feat: implement activate skill with conflict detection`
- `test: cover activation usecases`

---

### Fase 2 — Filesystem adapters

**Objetivo**: implementar leitura do repositório central, scanner de projetos e symlink manager. **Esta é a fase mais arriscada — symlinks têm armadilhas.**

#### 2.1. `SkillRepository` (filesystem)

- Lê `~/.skills-manager/skills/` (configurável via env `SKILLS_MANAGER_HOME`).
- Cada subdiretório com `SKILL.md` é uma skill.
- Parsea YAML frontmatter do `SKILL.md` pra extrair `name`, `description`.
- ID = SHA-256 dos primeiros bytes do path relativo (estável entre runs).
- Cache em memória, invalidado via watcher (`fsnotify`) — opcional na fase 2, pode ficar pra fase 7.

#### 2.2. `ProjectScanner`

- Recebe lista de "workspace roots" (ex: `~/dev/`).
- Walk recursivo com profundidade limitada (ex: 3 níveis) e respeitando `.gitignore`-like exclusions (`node_modules`, `.git/objects`, etc.).
- Detecta projeto por presença de `.git/`.
- Pra cada projeto, detecta agentes:
  - Claude: `.claude/` ou `CLAUDE.md` ou `AGENTS.md`.
  - Copilot: `.github/copilot-instructions.md` ou `.github/instructions/`.
- Retorna `[]ProjectCandidate` (não persiste — quem persiste é o caso de uso após confirmação do usuário).

#### 2.3. `SymlinkManager`

- Interface:
  ```go
  type SymlinkManager interface {
      EnsureLink(ctx context.Context, source, target string) error
      RemoveLink(ctx context.Context, target string) error
      IsManagedLink(ctx context.Context, target string) (bool, error) // verifica se aponta pra dentro do skills-manager home
  }
  ```
- Implementação única pra Linux + macOS via `os.Symlink`. Sem branching por SO, sem capability check.
- `EnsureLink` é idempotente:
  - Se target já existe e é symlink apontando pro lugar certo: no-op.
  - Se target existe e é symlink managed apontando pra lugar errado: remove e recria.
  - Se target existe mas **não é symlink** (é diretório/arquivo do usuário): retorna erro `ErrTargetNotManaged` — nunca sobrescreve nada que não seja managed.
  - Se target não existe: cria.
- `RemoveLink` só remove se `IsManagedLink == true` (segurança contra apagar pasta do usuário).
- `IsManagedLink` resolve o symlink e verifica se o destino está dentro de `~/.skills-manager/skills/`.
- Erros tipados pra UI tratar:
  - `ErrTargetNotManaged` — algo já existe lá e não é nosso.
  - `ErrSourceNotFound` — skill não existe mais no repositório central.
  - `ErrPermissionDenied` — sem permissão de escrita no path do agente.

#### 2.4. Testes

- `afero.MemMapFs` pros testes de `SkillRepository` e `ProjectScanner`.
- Testes de symlink usam `t.TempDir()` real (afero não suporta symlinks bem).
- CI roda matriz Linux + macOS — os dois precisam passar antes de merge.
- Atenção especial a um teste cobrindo macOS APFS case-insensitive: ativar "skill-X" e tentar criar symlink pra "Skill-X" deve detectar colisão.

**Critério de pronto**: 
- Skill repo lê 3 skills de mock e retorna entidades corretas.
- Scanner detecta projeto com `.git/` + `.claude/`.
- Symlink manager cria, detecta e remove links em Linux e macOS no CI.

**Commits**:
- `feat: add skill repository reading from central path`
- `feat: add project scanner with agent detection`
- `feat: add symlink manager`
- `test: cover filesystem adapters`

---

### Fase 3 — Agent adapters (Claude + Copilot)

**Objetivo**: implementar a interface `AgentAdapter` pros dois agentes do MVP.

#### 3.1. `ClaudeAdapter`

- **Global**: `~/.claude/skills/<skill-name>/` → symlink pra `~/.skills-manager/skills/<skill-name>/`.
- **Projeto**: `<projeto>/.claude/skills/<skill-name>/` → symlink análogo.
- `ApplyGlobal(skills)`: 
  1. Lista symlinks managed em `~/.claude/skills/`.
  2. Diff contra `skills` desejadas.
  3. Cria os faltantes, remove os sobrantes.
- `ApplyProject(project, skills)`: análogo, no path do projeto.
- `CapabilityCheck`: verifica se `~/.claude/` existe e é gravável; cria se não existir.

#### 3.2. `CopilotAdapter`

- **Não usa symlinks** — Copilot lê arquivo único.
- **Projeto**: gerencia `<projeto>/.github/copilot-instructions.md`.
- **Global**: Copilot **não tem conceito de global** oficialmente. Decisão: pro MVP, ativação global pra Copilot é **ignorada com warning na UI**. (Pós-MVP: poderíamos sincronizar com VSCode user settings, mas é frágil.)
- Estratégia de regeneração:
  1. Lê arquivo existente.
  2. Detecta blocos managed via marcadores:
     ```markdown
     <!-- skills-manager:start -->
     ... conteúdo gerado ...
     <!-- skills-manager:end -->
     ```
  3. Substitui o bloco managed pelo conteúdo concatenado das skills ativas.
  4. Conteúdo do usuário fora dos marcadores é preservado.
  5. Se arquivo não existe, cria com só o bloco managed.
- Conteúdo concatenado: pra cada skill ativa, insere `## <skill-name>` + corpo do `SKILL.md` (sem o frontmatter YAML).

#### 3.3. Registro de adapters via DI

- `samber/do` registra todos os adapters num `map[Agent]AgentAdapter`.
- Casos de uso resolvem o adapter pelo `Agent` da activation.

#### 3.4. Testes

- Adapter Claude: setup com `t.TempDir()` simulando `$HOME` e projeto. Verifica criação, idempotência, remoção.
- Adapter Copilot: testa preservação de conteúdo do usuário fora dos marcadores. Testa criação from-scratch. Testa regeneração com diff de skills.

**Critério de pronto**: `ApplyGlobal` e `ApplyProject` funcionam pros dois adapters; conflito é detectado corretamente quando skill é ativada nos dois escopos pro mesmo agente.

**Commits**:
- `feat: add claude agent adapter`
- `feat: add copilot agent adapter with managed block strategy`
- `test: cover agent adapters`

---

### Fase 4 — Persistência (SQLite)

**Objetivo**: implementar `ProjectRepository` e `ActivationRepository` em SQLite.

#### 4.1. Schema

```sql
CREATE TABLE projects (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    path        TEXT NOT NULL UNIQUE,
    added_at    TIMESTAMP NOT NULL
);

CREATE TABLE project_agents (
    project_id  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    agent       TEXT NOT NULL,
    PRIMARY KEY (project_id, agent)
);

CREATE TABLE activations (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    skill_id    TEXT NOT NULL,
    agent       TEXT NOT NULL,
    scope       TEXT NOT NULL CHECK (scope IN ('global', 'project')),
    project_id  TEXT REFERENCES projects(id) ON DELETE CASCADE,
    applied_at  TIMESTAMP NOT NULL,
    UNIQUE (skill_id, agent, scope, project_id)
);

CREATE INDEX idx_activations_lookup ON activations(skill_id, agent, project_id);
```

Constraint adicional aplicada via trigger ou check em código:
- Quando `scope = 'global'`, `project_id` deve ser NULL.
- Quando `scope = 'project'`, `project_id` não pode ser NULL.

#### 4.2. Migrations

- `internal/adapter/persistence/migrations/001_init.sql`.
- Runner simples no startup (`PRAGMA user_version`).
- Cada migration nova bumps `user_version`.

#### 4.3. Localização do DB

- `~/.skills-manager/registry.db`.
- Modo WAL ativado pra robustez.

#### 4.4. Testes

- DB em memória (`:memory:`) pra cada teste.
- Cobre CRUD + query de conflito.

**Critério de pronto**: repos passam contract tests definidos na fase 1; DB sobrevive restart do app.

**Commits**:
- `feat: add sqlite schema and migrations`
- `feat: implement project and activation repositories`
- `test: cover sqlite repositories`

---

### Fase 5 — Wails bindings

**Objetivo**: expor casos de uso pro frontend via bindings tipados.

#### 5.1. Estrutura

- `internal/binding/skills_binding.go`:
  ```go
  type SkillsBinding struct {
      listSkills usecase.ListSkills
  }

  func (b *SkillsBinding) List(ctx context.Context) ([]SkillDTO, error) { ... }
  ```
- DTOs separados das entidades de domínio (evita vazar internals).
- Bindings registrados em `app.go` no struct exposto pelo Wails.

#### 5.2. Bindings necessários

- `SkillsBinding.List()`
- `ProjectsBinding.List()`
- `ProjectsBinding.RegisterManual(path string)`
- `ProjectsBinding.ScanCandidates(roots []string)`
- `ProjectsBinding.ConfirmCandidate(candidate ProjectCandidateDTO)`
- `ProjectsBinding.Delete(id string)`
- `ActivationBinding.List(filter ActivationFilterDTO)`
- `ActivationBinding.Activate(req ActivateRequestDTO) (ActivateResultDTO, error)` — retorna `Conflict` se houver
- `ActivationBinding.Deactivate(id int64)`
- `ActivationBinding.ResolveConflict(req ResolveConflictRequestDTO)`
- `DoctorBinding.Run() (DoctorReportDTO, error)`

#### 5.3. Geração de tipos TS

- Wails gera types TS automaticamente em `frontend/wailsjs/go/`.
- Frontend importa via `infra/bindings.ts` (wrapper tipado, com tratamento de erro padronizado).

**Critério de pronto**: frontend chama `await window.go.binding.SkillsBinding.List()` e recebe array tipado.

**Commits**:
- `feat: add wails bindings for skills, projects and activations`

---

### Fase 6 — Frontend (React/TS + styled-components)

**Objetivo**: GUI completa pro MVP.

#### 6.1. Telas

1. **Skills Page** (`/skills`)
   - Lista de skills do repositório central.
   - Card por skill: nome, descrição, status (ativa onde?).
   - Filtro por agente / por estado (ativa global / ativa em algum projeto / inativa).

2. **Projects Page** (`/projects`)
   - Lista de projetos cadastrados.
   - Botão "Adicionar manualmente" (abre file picker via Wails).
   - Botão "Escanear workspace" (input de path raiz, lista candidatos, multi-select pra confirmar).

3. **Project Detail Page** (`/projects/:id`)
   - Info do projeto (path, agentes detectados).
   - Lista de skills com toggle por agente:
     - `[Claude: ON] [Copilot: -]` (— quando agente não existe no projeto).
   - Indicador de conflito quando aplicável.

4. **Doctor Page** (`/doctor`)
   - Roda validação: symlinks órfãos, skills no DB que não existem mais no FS, projetos com path inválido.
   - Botão "Reparar tudo".

#### 6.2. Modal de conflito

- Disparado quando `ActivationBinding.Activate` retorna `Conflict`.
- Mostra: skill, agente, ativação global existente, ativação proposta.
- Três botões: **Manter global**, **Override no projeto**, **Cancelar**.

#### 6.3. Componentes styled

- `<StyledCard>`, `<StyledToggle>`, `<StyledBadge>` (estado: active, conflict, disabled).
- Theme com tokens: `colors.surface`, `colors.conflict`, `colors.success`, etc.
- Dark mode opcional pós-MVP.

#### 6.4. Camadas (Container Pattern)

- **Types**: DTOs do backend re-exportados/refinados.
- **Infra**: `bindings.ts` + queries TanStack.
- **Pages**: composição de containers.
- **Routes**: TanStack Router file-based, com `Suspense` + `ErrorBoundary` em cada rota.
- **Components**: visuais puros, sem dependência de query.

**Critério de pronto**: fluxo end-to-end funciona — adicionar projeto, ativar skill, ver conflito, resolver, ver symlink criado no FS.

**Commits**:
- `feat: add skills page`
- `feat: add projects page with scan and manual add`
- `feat: add project detail with per-agent activation`
- `feat: add conflict resolution modal`
- `feat: add doctor page`

---

### Fase 7 — Polimento & doctor

**Objetivo**: robustez pra uso diário.

**Tarefas**:
- **Dry-run**: antes de aplicar uma ativação, mostra o que vai mudar no FS (lista de symlinks a criar/remover, arquivos a modificar).
- **File watcher** (`fsnotify`): se o usuário adicionar uma skill nova em `~/.skills-manager/skills/`, a UI reflete sem refresh manual.
- **Doctor avançado**:
  - Detecta symlinks managed quebrados (target não existe mais).
  - Detecta arquivos `copilot-instructions.md` cujo bloco managed foi corrompido.
  - Sugere ações de reparo.
- **Backup do registry** antes de migrations.
- **Logs estruturados** (slog) gravados em `~/.skills-manager/logs/`.
- **Tela de configurações**: workspace roots padrão, path do `SKILLS_MANAGER_HOME`, comportamento de conflito padrão.
- **Empty states e loading states** decentes em todas as telas.

**Critério de pronto**: usar o app por uma semana sem ter que abrir terminal pra desentortar nada.

**Commits**:
- `feat: add dry-run preview for activations`
- `feat: add filesystem watcher for skills repo`
- `feat: enhance doctor with broken symlink detection`
- `feat: add structured logging`
- `feat: add settings page`

---

## Riscos & armadilhas

### Case-sensitivity entre Linux e macOS
**Risco**: Linux ext4/btrfs é case-sensitive; macOS APFS por padrão é case-insensitive (preserva case mas não distingue). Skill chamada `pt-review` e `PT-Review` colidem no Mac mas são distintas no Linux. Um repositório de skills que funciona num Linux pode quebrar quando aberto no Mac.
**Mitigação**: validar nomes de skills no carregamento — rejeitar par de skills cujos nomes diferem só por case; documentar restrição no README; lint rule no doctor.

### Symlinks órfãos quando skill é renomeada/movida
**Risco**: usuário renomeia diretório de uma skill no repositório central; symlinks antigos ficam quebrados; activations no DB referenciam skill que não existe.
**Mitigação**: doctor detecta e oferece reparo; ID da skill é hash estável do path relativo (renomear = nova skill, antiga vira "missing" no doctor).

### Copilot bloco managed corrompido
**Risco**: usuário edita manualmente entre os marcadores; próxima regeneração apaga.
**Mitigação**: comentário claro no bloco managed avisando "do not edit"; doctor detecta divergências e avisa.

### Race conditions no filesystem
**Risco**: usuário abre dois projetos no Wails ao mesmo tempo, aplica ativações concorrentes.
**Mitigação**: lock por path (mutex em memória) durante operações de ativação; SQLite serializa naturalmente.

### Git no repositório central
**Risco**: usuário versiona `~/.skills-manager/skills/` com git; `.git/` aparece como skill.
**Mitigação**: scanner ignora diretórios começando com `.`.

### Path com espaços/unicode
**Risco**: path do projeto com espaços ou caracteres não-ASCII quebra symlink em algum SO.
**Mitigação**: testes com paths exóticos no CI.

### Tamanho do `copilot-instructions.md`
**Risco**: muitas skills ativas = arquivo gigante = Copilot trunca.
**Mitigação**: warning na UI quando ultrapassa N kb (limite documentado do Copilot).

---

## Pós-MVP

- **Suporte a Windows**: avaliar fallback de cópia (com watcher pra re-sincronizar quando source muda) ou exigir Developer Mode com tela explicativa. Decidir após validar o produto em Unix.
- **CLI complementar** (`skills-manager activate <skill> --project <path>`) pra automação.
- **Mais agentes**: Cursor (`.cursor/rules/`), Aider (`.aider.conf.yml`), Continue (`.continue/`).
- **Sync com repo git remoto**: skills-manager home apontando pra um clone, com pull/push integrado.
- **Templates de skills**: criar nova skill a partir de template via UI.
- **Skill packs**: agrupar skills relacionadas, ativar em conjunto.
- **Profile de ativação**: salvar conjuntos nomeados ("modo backend Go", "modo frontend React") e aplicar.
- **Export/import de configuração** entre máquinas.
- **Telemetria local** (sempre offline): quais skills mais ativadas, em quais projetos.
