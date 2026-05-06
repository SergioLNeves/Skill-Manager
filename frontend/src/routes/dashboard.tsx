import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { Layers, FolderOpen, Tag, Search } from 'lucide-react'
import { useAllSkills, useProjects, useCategories } from '@/infra/queries'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { AggregatedSkill } from '@/types'

export const Route = createFileRoute('/dashboard')({
  component: DashboardPage,
})

function StatCard({
  icon: Icon,
  label,
  value,
  loading,
}: {
  icon: React.ElementType
  label: string
  value: number
  loading: boolean
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{label}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        {loading ? (
          <div className="h-8 w-12 rounded bg-muted animate-pulse" />
        ) : (
          <p className="text-3xl font-bold">{value}</p>
        )}
      </CardContent>
    </Card>
  )
}

function SkillResult({ skill }: { skill: AggregatedSkill }) {
  const navigate = useNavigate()
  return (
    <button
      className="w-full text-left rounded-lg border border-border px-4 py-3 hover:bg-accent transition-colors"
      onClick={() => navigate({ to: '/skills/$name', params: { name: skill.name } })}
    >
      <div className="flex items-center justify-between gap-2">
        <span className="font-medium text-sm">{skill.name}</span>
        {skill.categoryName && (
          <span className="text-xs text-muted-foreground bg-muted rounded-full px-2 py-0.5 shrink-0">
            {skill.categoryName}
          </span>
        )}
      </div>
      {skill.description && skill.description !== skill.name && (
        <p className="text-xs text-muted-foreground mt-1 line-clamp-1">{skill.description}</p>
      )}
    </button>
  )
}

function DashboardPage() {
  const { data: skills = [], isLoading: loadingSkills } = useAllSkills()
  const { data: projects = [], isLoading: loadingProjects } = useProjects()
  const { data: categories = [], isLoading: loadingCategories } = useCategories()
  const [search, setSearch] = useState('')

  const query = search.trim().toLowerCase()
  const results: AggregatedSkill[] = query
    ? skills.filter(
        (s) =>
          s.name.toLowerCase().includes(query) ||
          s.description.toLowerCase().includes(query) ||
          s.categoryName?.toLowerCase().includes(query),
      )
    : []

  return (
    <div className="space-y-8 max-w-2xl">
      <div>
        <h1 className="text-2xl font-semibold">Dashboard</h1>
        <p className="text-sm text-muted-foreground mt-1">Visão geral do seu ambiente de skills.</p>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <StatCard
          icon={FolderOpen}
          label="Projetos"
          value={projects.length}
          loading={loadingProjects}
        />
        <StatCard
          icon={Layers}
          label="Skills"
          value={skills.length}
          loading={loadingSkills}
        />
        <StatCard
          icon={Tag}
          label="Categorias"
          value={categories.length}
          loading={loadingCategories}
        />
      </div>

      <div className="space-y-3">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
          <input
            type="text"
            placeholder="Buscar skills por nome, descrição ou categoria…"
            className="w-full rounded-md border border-input bg-background pl-9 pr-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>

        {query && (
          <div className="space-y-2">
            {results.length === 0 ? (
              <p className="text-sm text-muted-foreground px-1">Nenhuma skill encontrada para "{search}".</p>
            ) : (
              results.map((s) => <SkillResult key={s.name} skill={s} />)
            )}
          </div>
        )}
      </div>
    </div>
  )
}
