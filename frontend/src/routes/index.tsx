import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { Search, Copy, Trash2, Globe, FolderOpen } from 'lucide-react'
import { useAllSkills, useProjects, useCopySkill, useDeleteSkill } from '@/infra/queries'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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

  const filtered = skills.filter((s) => {
    const q = search.toLowerCase()
    return (
      s.name.toLowerCase().includes(q) ||
      s.description.toLowerCase().includes(q) ||
      s.projects.some((p) => p.name.toLowerCase().includes(q))
    )
  })

  if (isLoading) return <p className="text-muted-foreground">Carregando skills…</p>

  return (
    <div className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold">Skills</h1>
        <p className="text-sm text-muted-foreground mt-1">
          {skills.length} skill{skills.length !== 1 ? 's' : ''} encontrada{skills.length !== 1 ? 's' : ''}
        </p>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <input
          type="text"
          placeholder="Buscar por nome, descrição ou projeto…"
          className="w-full rounded-md border border-input bg-background pl-9 pr-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      {!filtered.length ? (
        <div className="text-center mt-16 text-muted-foreground">
          <p className="text-sm">Nenhuma skill encontrada.</p>
          {skills.length === 0 && (
            <p className="text-xs mt-1">
              Adicione projetos em <strong>Projects</strong> para ver as skills aqui.
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
              title="Copiar para projeto"
              onClick={onCopy}
            >
              <Copy className="h-3.5 w-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 hover:text-destructive"
              title="Excluir"
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
              <span className="font-medium text-foreground">Global</span>
            </div>
          )}
          {visibleProjects.map((p) => (
            <div key={p.id} className="flex items-center gap-1.5 text-xs text-muted-foreground">
              <FolderOpen className="h-3 w-3 shrink-0" />
              <span className="truncate">{p.name}</span>
            </div>
          ))}
          {overflow > 0 && (
            <p className="text-xs text-muted-foreground pl-[18px]">+{overflow} mais</p>
          )}
        </div>

        <div className="flex items-center gap-1.5 flex-wrap">
          {skill.isGlobal && (
            <Badge variant="outline" className="text-[10px] py-0 h-4">global</Badge>
          )}
          <Badge variant="secondary" className="text-[10px] py-0 h-4">
            {totalLocations} {totalLocations === 1 ? 'local' : 'locais'}
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
          <CardTitle className="text-base">Copiar skill para projetos</CardTitle>
          <p className="text-sm text-muted-foreground">
            Copia <strong>{skill.name}</strong> em <code>{destFolder}/</code> de cada projeto selecionado.
          </p>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <p className="text-xs font-medium text-muted-foreground">Destino (agente)</p>
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
            <p className="text-sm text-muted-foreground">Nenhum outro projeto registrado.</p>
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
                      {isDone && <p className="text-xs text-green-600">Copiado!</p>}
                    </div>
                  </label>
                )
              })}
            </div>
          )}
          <div className="flex gap-2 justify-end">
            <Button variant="outline" onClick={onClose}>Fechar</Button>
            {eligible.length > 0 && (
              <Button
                disabled={selected.length === 0 || copySkill.isPending}
                onClick={handleCopy}
              >
                {copySkill.isPending ? 'Copiando…' : `Copiar para ${selected.length} projeto${selected.length !== 1 ? 's' : ''}`}
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
    ...(skill.isGlobal ? [{ key: '__global__', label: 'Global', projectId: '' }] : []),
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
          <CardTitle className="text-base text-destructive">Excluir skill</CardTitle>
          <p className="text-sm text-muted-foreground">
            Selecione de quais locais excluir <strong>{skill.name}</strong>.
          </p>
        </CardHeader>
        <CardContent className="space-y-4">
          {allDone ? (
            <div className="space-y-3">
              <p className="text-sm text-green-600">
                Excluído. Backup salvo em <code>~/.skills-manager/trash/</code>.
              </p>
              <Button onClick={onClose} className="w-full">Fechar</Button>
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
                        {isDone && <p className="text-xs text-green-600">Excluído!</p>}
                      </div>
                    </label>
                  )
                })}
              </div>
              <p className="text-xs text-muted-foreground">
                Backup salvo em <code>~/.skills-manager/trash/</code> antes de excluir.
              </p>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={confirmed}
                  onChange={(e) => setConfirmed(e.target.checked)}
                />
                <span className="text-sm">Entendo que esta ação é irreversível no projeto</span>
              </label>
              <div className="flex gap-2 justify-end">
                <Button variant="outline" onClick={onClose}>Cancelar</Button>
                <Button
                  variant="destructive"
                  disabled={!confirmed || selectedLocations.length === 0 || deleteSkill.isPending}
                  onClick={handleDelete}
                >
                  {deleteSkill.isPending ? 'Excluindo…' : `Excluir ${selectedLocations.length} local${selectedLocations.length !== 1 ? 'is' : ''}`}
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
