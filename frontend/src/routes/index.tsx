import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { Search, Copy, Trash2, Globe, FolderOpen, GitFork } from 'lucide-react'
import { useAllSkills, useProjects, useCopySkill, useDeleteSkill } from '@/infra/queries'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { InstallFromGitHubModal } from '@/components/InstallFromGitHubModal'
import type { AggregatedSkill } from '@/types'

export const Route = createFileRoute('/')({
  component: SkillsPage,
})

function SkillsPage() {
  const { data: skills = [], isLoading } = useAllSkills()
  const { data: projects = [] } = useProjects()
  const [search, setSearch] = useState('')
  const [copyTarget, setCopyTarget] = useState<AggregatedSkill | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<AggregatedSkill | null>(null)
  const [showInstall, setShowInstall] = useState(false)

  const filtered = skills.filter((s) => {
    const q = search.toLowerCase()
    return (
      s.name.toLowerCase().includes(q) ||
      s.description.toLowerCase().includes(q) ||
      (s.categoryName ?? '').toLowerCase().includes(q) ||
      s.projects.some((p) => p.name.toLowerCase().includes(q))
    )
  })

  if (isLoading) return <p className="text-muted-foreground">Loading skills…</p>

  return (
    <div className="space-y-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">Skills</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {skills.length} skill{skills.length !== 1 ? 's' : ''} found
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="outline" size="sm" onClick={() => setShowInstall(true)}>
            <GitFork className="h-4 w-4 mr-1.5" />
            Add from GitHub
          </Button>
          <Link to="/categories" className="text-sm text-muted-foreground hover:text-foreground underline">
            Manage categories →
          </Link>
        </div>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <input
          type="text"
          placeholder="Search by name, description, category, or project…"
          className="w-full rounded-md border border-input bg-background pl-9 pr-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      {!filtered.length ? (
        <div className="text-center mt-16 text-muted-foreground">
          <p className="text-sm">No skills found.</p>
          {skills.length === 0 && (
            <p className="text-xs mt-1">
              Add projects in <strong>Projects</strong> to see skills here.
            </p>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {filtered.map((skill) => (
            <SkillCard
              key={skill.name}
              skill={skill}
              onCopy={(e) => { e.preventDefault(); setCopyTarget(skill) }}
              onDelete={(e) => { e.preventDefault(); setDeleteTarget(skill) }}
            />
          ))}
        </div>
      )}

      {copyTarget && (
        <CopyModal skill={copyTarget} projects={projects} onClose={() => setCopyTarget(null)} />
      )}
      {deleteTarget && (
        <DeleteModal skill={deleteTarget} onClose={() => setDeleteTarget(null)} />
      )}
      {showInstall && (
        <InstallFromGitHubModal onClose={() => setShowInstall(false)} />
      )}
    </div>
  )
}

function SkillCard({
  skill,
  onCopy,
  onDelete,
}: {
  skill: AggregatedSkill
  onCopy: (e: React.MouseEvent) => void
  onDelete: (e: React.MouseEvent) => void
}) {
  const totalLocations = (skill.isGlobal ? 1 : 0) + skill.projects.length
  const visibleProjects = skill.projects.slice(0, 3)
  const overflow = skill.projects.length - 3

  return (
    <Link to="/skills/$name" params={{ name: skill.name }} className="block group">
      <div className="rounded-lg border border-border hover:border-ring transition-colors p-4 space-y-3 h-full">
        <div className="flex items-start justify-between gap-2">
          <h3 className="text-sm font-semibold leading-tight group-hover:text-primary transition-colors">
            {skill.name}
          </h3>
          <div className="flex gap-1 shrink-0 -mt-1 -mr-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              title="Copy to project"
              onClick={onCopy}
            >
              <Copy className="h-3.5 w-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 hover:text-destructive"
              title="Delete"
              onClick={onDelete}
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>

        {skill.description && (
          <p className="text-xs text-muted-foreground line-clamp-2">{skill.description}</p>
        )}

        <div className="space-y-1">
          {skill.isGlobal && (
            <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <Globe className="h-3 w-3 shrink-0" />
              <span className="font-medium text-foreground">GitHub</span>
            </div>
          )}
          {visibleProjects.map((p) => (
            <div key={p.id} className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <FolderOpen className="h-3 w-3 shrink-0" />
              <span className="truncate">{p.name}</span>
            </div>
          ))}
          {overflow > 0 && (
            <p className="text-xs text-muted-foreground pl-[18px]">+{overflow} more</p>
          )}
        </div>

        <div className="flex items-center gap-1.5 flex-wrap">
          {skill.categoryName && (
            <Badge variant="outline" className="text-[10px] py-0 h-4">{skill.categoryName}</Badge>
          )}
          {skill.isGlobal && (
            <Badge variant="outline" className="text-[10px] py-0 h-4">github</Badge>
          )}
          <Badge variant="secondary" className="text-[10px] py-0 h-4">
            {totalLocations} {totalLocations === 1 ? 'location' : 'locations'}
          </Badge>
        </div>
      </div>
    </Link>
  )
}

function CopyModal({
  skill,
  projects,
  onClose,
}: {
  skill: AggregatedSkill
  projects: import('@/types').Project[]
  onClose: () => void
}) {
  const copySkill = useCopySkill()
  const [agent, setAgent] = useState<'claude' | 'copilot'>('claude')
  const [selected, setSelected] = useState<string[]>([])
  const [done, setDone] = useState<string[]>([])
  const [errors, setErrors] = useState<Record<string, string>>({})

  const alreadyHas = new Set(skill.projects.map((p) => p.id))
  const eligible = projects.filter((p) => !alreadyHas.has(p.id))
  const destFolder = agent === 'copilot' ? '.github/skills' : '.claude/skills'

  async function handleCopy() {
    for (const projectId of selected) {
      try {
        await copySkill.mutateAsync({
          skillId: skill.name,
          sourceProjectId: skill.projects[0]?.id ?? '',
          targetProjectId: projectId,
          agent,
        })
        setDone((d) => [...d, projectId])
      } catch (e) {
        setErrors((prev) => ({ ...prev, [projectId]: String(e) }))
      }
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <Card className="w-full max-w-md mx-4">
        <CardHeader>
          <CardTitle className="text-base">Copy skill to projects</CardTitle>
          <p className="text-sm text-muted-foreground">
            Copies <strong>{skill.name}</strong> into <code>{destFolder}/</code> of each selected project.
          </p>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <p className="text-xs font-medium text-muted-foreground">Target agent</p>
            <div className="flex gap-2">
              {(['claude', 'copilot'] as const).map((a) => (
                <button
                  key={a}
                  onClick={() => setAgent(a)}
                  className={`flex-1 rounded-md border px-3 py-1.5 text-sm font-medium transition-colors ${agent === a ? 'border-ring bg-accent' : 'border-border hover:bg-accent/50'}`}
                >
                  {a === 'claude' ? 'Claude (.claude/)' : 'Copilot (.github/)'}
                </button>
              ))}
            </div>
          </div>

          {eligible.length === 0 ? (
            <p className="text-sm text-muted-foreground">No other registered projects.</p>
          ) : (
            <div className="space-y-2 max-h-56 overflow-y-auto">
              {eligible.map((p) => {
                const isDone = done.includes(p.id)
                const err = errors[p.id]
                return (
                  <label
                    key={p.id}
                    className="flex items-center gap-3 rounded-md border border-border px-3 py-2 cursor-pointer hover:bg-accent"
                  >
                    <input
                      type="checkbox"
                      disabled={isDone}
                      checked={selected.includes(p.id) || isDone}
                      onChange={(e) =>
                        setSelected(
                          e.target.checked
                            ? [...selected, p.id]
                            : selected.filter((id) => id !== p.id),
                        )
                      }
                    />
                    <div className="flex-1">
                      <p className="text-sm font-medium">{p.name}</p>
                      {err && <p className="text-xs text-destructive">{err}</p>}
                      {isDone && <p className="text-xs text-green-600">Copied</p>}
                    </div>
                  </label>
                )
              })}
            </div>
          )}
          <div className="flex gap-2 justify-end">
            <Button variant="outline" onClick={onClose}>Close</Button>
            {eligible.length > 0 && (
              <Button
                disabled={selected.length === 0 || copySkill.isPending}
                onClick={handleCopy}
              >
                {copySkill.isPending ? 'Copying…' : `Copy to ${selected.length} project${selected.length !== 1 ? 's' : ''}`}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function DeleteModal({
  skill,
  onClose,
}: {
  skill: AggregatedSkill
  onClose: () => void
}) {
  const deleteSkill = useDeleteSkill()
  const [selectedLocations, setSelectedLocations] = useState<string[]>([])
  const [done, setDone] = useState<string[]>([])
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [confirmed, setConfirmed] = useState(false)

  const locations = [
    ...(skill.isGlobal ? [{ key: '__global__', label: 'GitHub (managed)', projectId: '' }] : []),
    ...skill.projects.map((p) => ({ key: p.id, label: p.name, projectId: p.id })),
  ]

  async function handleDelete() {
    for (const loc of locations.filter((l) => selectedLocations.includes(l.key))) {
      try {
        await deleteSkill.mutateAsync({ skillId: skill.name, projectId: loc.projectId })
        setDone((d) => [...d, loc.key])
      } catch (e) {
        setErrors((prev) => ({ ...prev, [loc.key]: String(e) }))
      }
    }
  }

  const allDone = selectedLocations.length > 0 && selectedLocations.every((k) => done.includes(k))

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <Card className="w-full max-w-sm mx-4">
        <CardHeader>
          <CardTitle className="text-base text-destructive">Delete skill</CardTitle>
          <p className="text-sm text-muted-foreground">
            Select the locations to delete <strong>{skill.name}</strong> from.
          </p>
        </CardHeader>
        <CardContent className="space-y-4">
          {allDone ? (
            <div className="space-y-3">
              <p className="text-sm text-green-600">
                Deleted. Backup saved at <code>~/.skills-manager/trash/</code>.
              </p>
              <Button onClick={onClose} className="w-full">Close</Button>
            </div>
          ) : (
            <>
              <div className="space-y-2 max-h-48 overflow-y-auto">
                {locations.map((loc) => {
                  const isDone = done.includes(loc.key)
                  const err = errors[loc.key]
                  return (
                    <label
                      key={loc.key}
                      className="flex items-center gap-3 rounded-md border border-border px-3 py-2 cursor-pointer hover:bg-accent"
                    >
                      <input
                        type="checkbox"
                        disabled={isDone}
                        checked={selectedLocations.includes(loc.key) || isDone}
                        onChange={(e) =>
                          setSelectedLocations(
                            e.target.checked
                              ? [...selectedLocations, loc.key]
                              : selectedLocations.filter((k) => k !== loc.key),
                          )
                        }
                      />
                      <div className="flex-1">
                        <p className="text-sm font-medium">{loc.label}</p>
                        {err && <p className="text-xs text-destructive">{err}</p>}
                        {isDone && <p className="text-xs text-green-600">Deleted</p>}
                      </div>
                    </label>
                  )
                })}
              </div>
              <p className="text-xs text-muted-foreground">
                Backup saved at <code>~/.skills-manager/trash/</code> before deletion.
              </p>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={confirmed}
                  onChange={(e) => setConfirmed(e.target.checked)}
                />
                <span className="text-sm">I understand this action is irreversible.</span>
              </label>
              <div className="flex gap-2 justify-end">
                <Button variant="outline" onClick={onClose}>Cancel</Button>
                <Button
                  variant="destructive"
                  disabled={!confirmed || selectedLocations.length === 0 || deleteSkill.isPending}
                  onClick={handleDelete}
                >
                  {deleteSkill.isPending ? 'Deleting…' : `Delete ${selectedLocations.length} location${selectedLocations.length !== 1 ? 's' : ''}`}
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
