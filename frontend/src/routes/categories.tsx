import { createFileRoute, Link } from '@tanstack/react-router'
import { useMemo, useState } from 'react'
import { FolderPlus, Tag, ArrowLeft, Plus, Pencil, Trash2, X, Check, Users } from 'lucide-react'
import {
  useCategories,
  useCreateCategory,
  useUpdateCategory,
  useDeleteCategory,
  useAllSkills,
  useProjects,
  useAssociateProjectCategory,
  useAssignSkillCategory,
} from '@/infra/queries'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { AggregatedSkill, Category, Project } from '@/types'

export const Route = createFileRoute('/categories')({
  component: CategoriesPage,
})

function CategoriesPage() {
  const { data: categories = [], isLoading } = useCategories()
  const { data: skills = [] } = useAllSkills()
  const { data: projects = [] } = useProjects()
  const createCategory = useCreateCategory()
  const [showCreate, setShowCreate] = useState(false)
  const [addTarget, setAddTarget] = useState<Category | null>(null)
  const [manageTarget, setManageTarget] = useState<Category | null>(null)

  const skillsByCategory = useMemo(() => {
    const map: Record<number, AggregatedSkill[]> = {}
    for (const s of skills) {
      if (s.categoryId != null) {
        ;(map[s.categoryId] ??= []).push(s)
      }
    }
    return map
  }, [skills])

  if (isLoading) return <p className="text-muted-foreground">Loading…</p>

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <Link to="/" className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-3">
          <ArrowLeft className="h-3.5 w-3.5" /> Skills
        </Link>
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold">Categories</h1>
            <p className="text-sm text-muted-foreground mt-1">
              Organise skills by topic. Associating a category with a project copies all its
              skills there — new skills added to the category are automatically propagated.
            </p>
          </div>
          <Button size="sm" onClick={() => setShowCreate(true)}>
            <Plus className="h-4 w-4 mr-1.5" />
            New category
          </Button>
        </div>
      </div>

      {showCreate && (
        <CreateCategoryForm
          onSave={async (name, desc) => {
            await createCategory.mutateAsync({ name, description: desc })
            setShowCreate(false)
          }}
          onCancel={() => setShowCreate(false)}
          isPending={createCategory.isPending}
        />
      )}

      {categories.length === 0 && !showCreate ? (
        <div className="rounded-lg border border-dashed border-border p-12 text-center text-muted-foreground">
          <Tag className="h-10 w-10 mx-auto mb-3 opacity-30" />
          <p className="text-sm">No categories yet. Create one to start organising skills.</p>
        </div>
      ) : (
        <div className="space-y-3">
          {categories.map((cat) => (
            <CategoryRow
              key={cat.id}
              category={cat}
              skills={skillsByCategory[cat.id] ?? []}
              onAddToProject={() => setAddTarget(cat)}
              onManageSkills={() => setManageTarget(cat)}
            />
          ))}
        </div>
      )}

      {addTarget && (
        <AssociateModal
          category={addTarget}
          skills={skillsByCategory[addTarget.id] ?? []}
          projects={projects}
          onClose={() => setAddTarget(null)}
        />
      )}

      {manageTarget && (
        <ManageSkillsModal
          category={manageTarget}
          allSkills={skills}
          currentSkills={skillsByCategory[manageTarget.id] ?? []}
          onClose={() => setManageTarget(null)}
        />
      )}
    </div>
  )
}

function CreateCategoryForm({
  onSave,
  onCancel,
  isPending,
}: {
  onSave: (name: string, desc: string) => void
  onCancel: () => void
  isPending: boolean
}) {
  const [name, setName] = useState('')
  const [desc, setDesc] = useState('')
  return (
    <div className="rounded-lg border border-border p-4 space-y-3 bg-muted/20">
      <p className="text-sm font-medium">New category</p>
      <input
        autoFocus
        type="text"
        placeholder="Name"
        value={name}
        onChange={(e) => setName(e.target.value)}
        onKeyDown={(e) => e.key === 'Enter' && name.trim() && onSave(name.trim(), desc.trim())}
        className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
      />
      <input
        type="text"
        placeholder="Description (optional)"
        value={desc}
        onChange={(e) => setDesc(e.target.value)}
        className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
      />
      <div className="flex gap-2 justify-end">
        <Button variant="outline" size="sm" onClick={onCancel} disabled={isPending}>Cancel</Button>
        <Button size="sm" disabled={!name.trim() || isPending} onClick={() => onSave(name.trim(), desc.trim())}>
          {isPending ? 'Saving…' : 'Create'}
        </Button>
      </div>
    </div>
  )
}

function CategoryRow({
  category,
  skills,
  onAddToProject,
  onManageSkills,
}: {
  category: Category
  skills: AggregatedSkill[]
  onAddToProject: () => void
  onManageSkills: () => void
}) {
  const updateCategory = useUpdateCategory()
  const deleteCategory = useDeleteCategory()
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState(category.name)
  const [desc, setDesc] = useState(category.description)

  async function saveEdit() {
    await updateCategory.mutateAsync({ id: category.id, name: name.trim(), description: desc.trim() })
    setEditing(false)
  }

  async function handleDelete() {
    if (!confirm(`Delete category "${category.name}"? Skills will be unassigned but not deleted.`)) return
    await deleteCategory.mutateAsync(category.id)
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-4">
          {editing ? (
            <div className="flex-1 space-y-2">
              <input
                autoFocus
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
              <input
                value={desc}
                onChange={(e) => setDesc(e.target.value)}
                placeholder="Description"
                className="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              />
              <div className="flex gap-1.5">
                <Button size="sm" onClick={saveEdit} disabled={!name.trim() || updateCategory.isPending}>
                  <Check className="h-3.5 w-3.5 mr-1" /> Save
                </Button>
                <Button size="sm" variant="ghost" onClick={() => { setEditing(false); setName(category.name); setDesc(category.description) }}>
                  <X className="h-3.5 w-3.5 mr-1" /> Cancel
                </Button>
              </div>
            </div>
          ) : (
            <CardTitle className="text-base flex items-center gap-2">
              <Tag className="h-4 w-4 shrink-0" />
              {category.name}
              <Badge variant="secondary" className="text-[10px] py-0 h-4 ml-1">
                {skills.length} skill{skills.length !== 1 ? 's' : ''}
              </Badge>
            </CardTitle>
          )}

          {!editing && (
            <div className="flex gap-1 shrink-0 items-center">
              <Button variant="ghost" size="icon" className="h-7 w-7" title="Manage skills" onClick={onManageSkills}>
                <Users className="h-3.5 w-3.5" />
              </Button>
              <Button variant="ghost" size="icon" className="h-7 w-7" title="Edit" onClick={() => setEditing(true)}>
                <Pencil className="h-3.5 w-3.5" />
              </Button>
              <Button variant="ghost" size="icon" className="h-7 w-7 hover:text-destructive" title="Delete" onClick={handleDelete}>
                <Trash2 className="h-3.5 w-3.5" />
              </Button>
              <Button variant="outline" size="sm" onClick={onAddToProject} className="ml-1">
                <FolderPlus className="h-3.5 w-3.5 mr-1.5" />
                Add to project
              </Button>
            </div>
          )}
        </div>

        {!editing && category.description && (
          <p className="text-xs text-muted-foreground mt-1">{category.description}</p>
        )}
      </CardHeader>

      <CardContent className="pt-0">
        {skills.length === 0 ? (
          <p className="text-xs text-muted-foreground italic">
            No skills assigned yet.{' '}
            <button className="underline hover:text-foreground" onClick={onManageSkills}>Add skills</button>
          </p>
        ) : (
          <div className="flex flex-wrap gap-1.5">
            {skills.map((s) => (
              <Link
                key={s.name}
                to="/skills/$name"
                params={{ name: s.name }}
                className="text-xs rounded-md border border-border px-2 py-1 hover:bg-accent transition-colors"
              >
                {s.name}
              </Link>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function ManageSkillsModal({
  category,
  allSkills,
  currentSkills,
  onClose,
}: {
  category: Category
  allSkills: AggregatedSkill[]
  currentSkills: AggregatedSkill[]
  onClose: () => void
}) {
  const assign = useAssignSkillCategory()
  const currentNames = useMemo(() => new Set(currentSkills.map((s) => s.name)), [currentSkills])
  const [pending, setPending] = useState<Record<string, boolean>>({})
  const [search, setSearch] = useState('')

  const filtered = allSkills.filter((s) =>
    s.name.toLowerCase().includes(search.toLowerCase()) ||
    s.description.toLowerCase().includes(search.toLowerCase())
  )

  async function toggle(skill: AggregatedSkill) {
    const isIn = currentNames.has(skill.name)
    const primaryPath = skill.isGlobal ? skill.globalPath : (skill.projects[0]?.skillPath ?? '')
    setPending((p) => ({ ...p, [skill.name]: true }))
    try {
      await assign.mutateAsync({
        skillName: skill.name,
        skillPath: primaryPath,
        categoryId: isIn ? null : category.id,
      })
    } finally {
      setPending((p) => ({ ...p, [skill.name]: false }))
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <Card className="w-full max-w-md mx-4 flex flex-col max-h-[80vh]">
        <CardHeader className="shrink-0">
          <CardTitle className="text-base">
            Manage skills — <span className="text-muted-foreground font-normal">{category.name}</span>
          </CardTitle>
          <p className="text-xs text-muted-foreground">
            Toggle skills to add or remove them from this category.
          </p>
          <input
            autoFocus
            type="text"
            placeholder="Search skills…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="mt-2 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </CardHeader>

        <CardContent className="flex-1 overflow-y-auto space-y-1 py-0 pb-4">
          {filtered.length === 0 ? (
            <p className="text-sm text-muted-foreground py-4 text-center">No skills found.</p>
          ) : (
            filtered.map((skill) => {
              const isIn = currentNames.has(skill.name)
              const isLoading = pending[skill.name]
              return (
                <button
                  key={skill.name}
                  disabled={isLoading}
                  onClick={() => toggle(skill)}
                  className={`w-full flex items-center gap-3 rounded-md border px-3 py-2.5 text-left transition-colors ${
                    isIn
                      ? 'border-ring bg-accent'
                      : 'border-border hover:bg-accent/50'
                  } ${isLoading ? 'opacity-50 cursor-wait' : 'cursor-pointer'}`}
                >
                  <div className={`h-4 w-4 shrink-0 rounded border flex items-center justify-center transition-colors ${isIn ? 'bg-primary border-primary' : 'border-muted-foreground'}`}>
                    {isIn && <Check className="h-3 w-3 text-primary-foreground" />}
                  </div>
                  <div className="min-w-0">
                    <p className="text-sm font-medium truncate">{skill.name}</p>
                    {skill.description && (
                      <p className="text-xs text-muted-foreground truncate">{skill.description}</p>
                    )}
                    {skill.categoryName && skill.categoryName !== category.name && (
                      <p className="text-xs text-amber-600">Currently in: {skill.categoryName}</p>
                    )}
                  </div>
                </button>
              )
            })
          )}
        </CardContent>

        <div className="shrink-0 border-t p-4">
          <Button variant="outline" className="w-full" onClick={onClose}>Done</Button>
        </div>
      </Card>
    </div>
  )
}

function AssociateModal({
  category,
  skills,
  projects,
  onClose,
}: {
  category: Category
  skills: AggregatedSkill[]
  projects: Project[]
  onClose: () => void
}) {
  const associate = useAssociateProjectCategory()
  const [agent, setAgent] = useState<'claude' | 'copilot'>('claude')
  const [projectId, setProjectId] = useState('')

  const destFolder = agent === 'copilot' ? '.github/skills' : '.claude/skills'

  async function handleAssociate() {
    if (!projectId) return
    await associate.mutateAsync({ projectId, categoryId: category.id, agent })
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <Card className="w-full max-w-md mx-4">
        <CardHeader>
          <CardTitle className="text-base">Associate category with project</CardTitle>
          <p className="text-sm text-muted-foreground">
            Links <strong>{category.name}</strong> ({skills.length} skill{skills.length !== 1 ? 's' : ''}) to a
            project. All current skills are copied into <code>{destFolder}/</code> and any skill
            added to this category later will be automatically copied too.
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

          {projects.length === 0 ? (
            <p className="text-sm text-muted-foreground">No registered projects.</p>
          ) : (
            <div className="space-y-1.5">
              <p className="text-xs font-medium text-muted-foreground">Target project</p>
              <select
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                value={projectId}
                onChange={(e) => setProjectId(e.target.value)}
              >
                <option value="">Select a project…</option>
                {projects.map((p) => (
                  <option key={p.id} value={p.id}>{p.name}</option>
                ))}
              </select>
            </div>
          )}

          <div className="flex gap-2 justify-end">
            <Button variant="outline" onClick={onClose}>Cancel</Button>
            <Button disabled={!projectId || associate.isPending} onClick={handleAssociate}>
              {associate.isPending ? 'Associating…' : 'Associate'}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
