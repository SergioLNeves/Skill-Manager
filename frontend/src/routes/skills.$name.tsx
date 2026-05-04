import { createFileRoute, Link } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkFrontmatter from 'remark-frontmatter'
import { ArrowLeft, Copy, Trash2, Globe, FolderOpen } from 'lucide-react'
import { useAllSkills, useProjects } from '@/infra/queries'
import { api } from '@/infra/bindings'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import type { AggregatedSkill } from '@/types'

export const Route = createFileRoute('/skills/$name')({
  component: SkillDetailPage,
})

function SkillDetailPage() {
  const { name } = Route.useParams()
  const { data: skills = [], isLoading } = useAllSkills()

  if (isLoading) return <p className="text-muted-foreground">Carregando…</p>

  const skill = skills.find((s) => s.name === name)
  if (!skill) {
    return (
      <div className="space-y-4">
        <Link to="/" className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-4 w-4" /> Voltar para Skills
        </Link>
        <p className="text-sm text-muted-foreground">Skill <strong>{name}</strong> não encontrada.</p>
      </div>
    )
  }

  return <SkillView skill={skill} />
}

function SkillView({ skill }: { skill: AggregatedSkill }) {
  const primaryPath = skill.isGlobal ? skill.globalPath : skill.projects[0]?.skillPath
  const contentPath = primaryPath ? `${primaryPath}/SKILL.md` : ''

  const [content, setContent] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!contentPath) return
    setLoading(true)
    setContent(null)
    setError(null)
    api.readSkillContent(contentPath)
      .then((t) => { setContent(t); setLoading(false) })
      .catch(() => { setError('Não foi possível ler o arquivo.'); setLoading(false) })
  }, [contentPath])

  const totalLocations = (skill.isGlobal ? 1 : 0) + skill.projects.length

  return (
    <div className="max-w-3xl mx-auto space-y-8">
      {/* Header */}
      <div className="space-y-4">
        <Link to="/" className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground w-fit">
          <ArrowLeft className="h-4 w-4" /> Voltar para Skills
        </Link>

        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <h1 className="text-2xl font-semibold">{skill.name}</h1>
            {skill.description && (
              <p className="text-sm text-muted-foreground">{skill.description}</p>
            )}
          </div>
          <div className="flex gap-2 shrink-0">
            <Button variant="outline" size="sm">
              <Copy className="h-3.5 w-3.5 mr-1.5" /> Copiar
            </Button>
            <Button variant="outline" size="sm" className="hover:text-destructive hover:border-destructive">
              <Trash2 className="h-3.5 w-3.5 mr-1.5" /> Excluir
            </Button>
          </div>
        </div>

        {/* Badges */}
        <div className="flex flex-wrap gap-1.5">
          {skill.isGlobal && (
            <Badge variant="outline" className="gap-1">
              <Globe className="h-3 w-3" /> Global
            </Badge>
          )}
          <Badge variant="secondary">
            {totalLocations} {totalLocations === 1 ? 'local' : 'locais'}
          </Badge>
        </div>
      </div>

      {/* Locations */}
      {(skill.isGlobal || skill.projects.length > 0) && (
        <div className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Localizações</p>
          <div className="space-y-2">
            {skill.isGlobal && (
              <div className="flex items-center gap-2 text-sm rounded-md border border-border px-3 py-2">
                <Globe className="h-4 w-4 text-muted-foreground shrink-0" />
                <div className="min-w-0">
                  <p className="font-medium">Global</p>
                  <p className="text-xs text-muted-foreground font-mono truncate">{skill.globalPath}</p>
                </div>
              </div>
            )}
            {skill.projects.map((p) => (
              <Link
                key={p.id}
                to="/projects/$id"
                params={{ id: p.id }}
                className="flex items-center gap-2 text-sm rounded-md border border-border px-3 py-2 hover:bg-accent transition-colors"
              >
                <FolderOpen className="h-4 w-4 text-muted-foreground shrink-0" />
                <div className="min-w-0">
                  <p className="font-medium">{p.name}</p>
                  <p className="text-xs text-muted-foreground font-mono truncate">{p.path}</p>
                </div>
              </Link>
            ))}
          </div>
        </div>
      )}

      {/* Markdown content */}
      <div className="space-y-2">
        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Conteúdo</p>
        {loading && <p className="text-sm text-muted-foreground">Carregando…</p>}
        {error && <p className="text-sm text-destructive">{error}</p>}
        {content && (
          <article className="prose prose-sm dark:prose-invert max-w-none prose-pre:bg-muted prose-pre:border prose-pre:border-border prose-code:text-foreground prose-headings:text-foreground prose-p:text-foreground/90 prose-li:text-foreground/90 prose-strong:text-foreground">
            <ReactMarkdown remarkPlugins={[remarkFrontmatter, remarkGfm]}>{content}</ReactMarkdown>
          </article>
        )}
      </div>
    </div>
  )
}
