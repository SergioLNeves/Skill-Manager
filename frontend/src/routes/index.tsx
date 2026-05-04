import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { Search, Copy, Trash2, ChevronDown, ChevronRight, BookOpen } from 'lucide-react'
import { useAllSkills, useProjects, useCopySkill, useDeleteSkill } from '@/infra/queries'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { Skill, Project } from '@/types'

export const Route = createFileRoute('/')({
  component: SkillsPage,
})

function SkillsPage() {
  const { data: skills = [], isLoading } = useAllSkills()
  const { data: projects = [] } = useProjects()
  const [search, setSearch] = useState('')
  const [selectedSkill, setSelectedSkill] = useState<Skill | null>(null)
  const [copyTarget, setCopyTarget] = useState<Skill | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Skill | null>(null)

  const filtered = skills.filter(
    (s) =>
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.description.toLowerCase().includes(search.toLowerCase()) ||
      s.ownerProjectName.toLowerCase().includes(search.toLowerCase()),
  )

  if (isLoading) return <p className="text-muted-foreground">Carregando skills…</p>

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Skills</h1>
        <p className="text-sm text-muted-foreground mt-1">
          {skills.length} skill{skills.length !== 1 ? 's' : ''} encontrada{skills.length !== 1 ? 's' : ''} em todos os projetos
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
              Adicione projetos em <strong>Projects</strong> para ver as skills deles aqui.
            </p>
          )}
        </div>
      ) : (
        <div className="grid gap-3">
          {filtered.map((skill) => (
            <SkillRow
              key={skill.id}
              skill={skill}
              isExpanded={selectedSkill?.id === skill.id}
              onToggleExpand={() =>
                setSelectedSkill(selectedSkill?.id === skill.id ? null : skill)
              }
              onCopy={() => setCopyTarget(skill)}
              onDelete={() => setDeleteTarget(skill)}
            />
          ))}
        </div>
      )}

      {copyTarget && (
        <CopyModal
          skill={copyTarget}
          projects={projects}
          onClose={() => setCopyTarget(null)}
        />
      )}

      {deleteTarget && (
        <DeleteModal
          skill={deleteTarget}
          onClose={() => setDeleteTarget(null)}
        />
      )}
    </div>
  )
}

function SkillRow({
  skill,
  isExpanded,
  onToggleExpand,
  onCopy,
  onDelete,
}: {
  skill: Skill
  isExpanded: boolean
  onToggleExpand: () => void
  onCopy: () => void
  onDelete: () => void
}) {
  return (
    <div className="rounded-lg border border-border">
      <div className="flex items-center gap-3 px-4 py-3">
        <button
          onClick={onToggleExpand}
          className="shrink-0 text-muted-foreground hover:text-foreground"
        >
          {isExpanded ? (
            <ChevronDown className="h-4 w-4" />
          ) : (
            <ChevronRight className="h-4 w-4" />
          )}
        </button>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-sm font-medium">{skill.name}</span>
            {skill.ownerProjectName && (
              <Badge variant="secondary" className="text-xs shrink-0">
                {skill.ownerProjectName}
              </Badge>
            )}
            {skill.source === 'global' && (
              <Badge variant="outline" className="text-xs shrink-0">global</Badge>
            )}
          </div>
          {skill.description && (
            <p className="text-xs text-muted-foreground mt-0.5 truncate">{skill.description}</p>
          )}
        </div>
        <div className="flex gap-1 shrink-0">
          <Button variant="ghost" size="icon" title="Ver conteúdo" onClick={onToggleExpand}>
            <BookOpen className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" title="Copiar para projeto" onClick={onCopy}>
            <Copy className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            title="Excluir"
            className="hover:text-destructive"
            onClick={onDelete}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>
      {isExpanded && <SkillContent skill={skill} />}
    </div>
  )
}

function SkillContent({ skill }: { skill: Skill }) {
  const [content, setContent] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (!content && !loading && !error) {
    setLoading(true)
    fetch(`file://${skill.path}`)
      .catch(() => {
        // file:// not available in Wails — read via a different strategy
        // Show path so user knows where the file is
        setContent(`# ${skill.name}\n\n${skill.description}\n\n📁 \`${skill.path}\``)
        setLoading(false)
      })
      .then((res) => {
        if (res) {
          res.text().then((t) => { setContent(t); setLoading(false) })
        }
      })
      .catch(() => {
        setError('Não foi possível ler o arquivo.')
        setLoading(false)
      })
  }

  return (
    <div className="border-t border-border px-4 py-3 bg-muted/30">
      <p className="text-xs text-muted-foreground font-mono mb-2">{skill.path}</p>
      {loading && <p className="text-xs text-muted-foreground">Carregando…</p>}
      {error && <p className="text-xs text-destructive">{error}</p>}
      {content && (
        <pre className="text-xs whitespace-pre-wrap font-mono text-foreground/80 max-h-64 overflow-y-auto">
          {content}
        </pre>
      )}
    </div>
  )
}

function CopyModal({
  skill,
  projects,
  onClose,
}: {
  skill: Skill
  projects: Project[]
  onClose: () => void
}) {
  const copySkill = useCopySkill()
  const [selected, setSelected] = useState<string[]>([])
  const [done, setDone] = useState<string[]>([])
  const [errors, setErrors] = useState<Record<string, string>>({})

  const eligible = projects.filter((p) => p.id !== skill.ownerProjectId)

  async function handleCopy() {
    for (const projectId of selected) {
      try {
        await copySkill.mutateAsync({
          skillId: skill.id,
          sourceProjectId: skill.ownerProjectId,
          targetProjectId: projectId,
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
            Cria uma cópia física de <strong>{skill.name}</strong> em <code>skills/</code> de cada projeto selecionado.
          </p>
        </CardHeader>
        <CardContent className="space-y-4">
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
  skill: Skill
  onClose: () => void
}) {
  const deleteSkill = useDeleteSkill()
  const [confirmed, setConfirmed] = useState(false)
  const [done, setDone] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleDelete() {
    try {
      await deleteSkill.mutateAsync({ skillId: skill.id, projectId: skill.ownerProjectId })
      setDone(true)
    } catch (e) {
      setError(String(e))
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <Card className="w-full max-w-sm mx-4">
        <CardHeader>
          <CardTitle className="text-base text-destructive">Excluir skill</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {done ? (
            <div className="space-y-3">
              <p className="text-sm text-green-600">
                Skill excluída. Uma cópia de segurança foi salva em <code>~/.skills-manager/trash/</code>.
              </p>
              <Button onClick={onClose} className="w-full">Fechar</Button>
            </div>
          ) : (
            <>
              <p className="text-sm">
                Isso vai <strong>deletar permanentemente</strong> os arquivos de{' '}
                <strong>{skill.name}</strong> do disco.
              </p>
              <p className="text-xs text-muted-foreground">
                Uma cópia de segurança será salva em <code>~/.skills-manager/trash/</code> antes de excluir.
              </p>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={confirmed}
                  onChange={(e) => setConfirmed(e.target.checked)}
                />
                <span className="text-sm">Entendo que esta ação é irreversível no projeto</span>
              </label>
              {error && <p className="text-xs text-destructive">{error}</p>}
              <div className="flex gap-2 justify-end">
                <Button variant="outline" onClick={onClose}>Cancelar</Button>
                <Button
                  variant="destructive"
                  disabled={!confirmed || deleteSkill.isPending}
                  onClick={handleDelete}
                >
                  {deleteSkill.isPending ? 'Excluindo…' : 'Excluir'}
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
