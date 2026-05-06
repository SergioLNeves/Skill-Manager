import { createFileRoute, Link } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import remarkFrontmatter from 'remark-frontmatter'
import { ArrowLeft, Globe, Tag } from 'lucide-react'
import { useAllSkills } from '@/infra/queries'
import { api } from '@/infra/bindings'
import { Badge } from '@/components/ui/badge'
import type { AggregatedSkill } from '@/types'

export const Route = createFileRoute('/skills/$name')({
  component: SkillDetailPage,
})

function SkillDetailPage() {
  const { name } = Route.useParams()
  const { data: skills = [], isLoading } = useAllSkills()

  if (isLoading) return <p className="text-muted-foreground">Loading…</p>

  const skill = skills.find((s) => s.name === name)
  if (!skill) {
    return (
      <div className="space-y-4">
        <Link to="/" className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-4 w-4" /> Back to Skills
        </Link>
        <p className="text-sm text-muted-foreground">Skill <strong>{name}</strong> not found.</p>
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
      .catch(() => { setError('Could not read the file.'); setLoading(false) })
  }, [contentPath])

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <Link to="/" className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground w-fit">
        <ArrowLeft className="h-4 w-4" /> Back to Skills
      </Link>

      <div className="space-y-2">
        <h1 className="text-2xl font-semibold">{skill.name}</h1>
        {skill.description && (
          <p className="text-sm text-muted-foreground">{skill.description}</p>
        )}
        <div className="flex flex-wrap gap-1.5">
          {skill.categoryName && (
            <Badge variant="outline" className="gap-1">
              <Tag className="h-3 w-3" /> {skill.categoryName}
            </Badge>
          )}
          {skill.isGlobal && (
            <Badge variant="outline" className="gap-1">
              <Globe className="h-3 w-3" /> GitHub
            </Badge>
          )}
        </div>
      </div>

      <div className="space-y-2">
        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Content</p>
        {loading && <p className="text-sm text-muted-foreground">Loading…</p>}
        {error && <p className="text-sm text-destructive">{error}</p>}
        {!loading && !error && !content && (
          <p className="text-sm text-muted-foreground italic">No SKILL.md found.</p>
        )}
        {content && (
          <article className="prose prose-sm dark:prose-invert max-w-none prose-pre:bg-muted prose-pre:border prose-pre:border-border prose-code:text-foreground prose-headings:text-foreground prose-p:text-foreground/90 prose-li:text-foreground/90 prose-strong:text-foreground">
            <ReactMarkdown remarkPlugins={[remarkFrontmatter, remarkGfm]}>{content}</ReactMarkdown>
          </article>
        )}
      </div>
    </div>
  )
}
