import { createFileRoute, useParams } from '@tanstack/react-router'
import { useState } from 'react'
import { ArrowLeft, RotateCcw } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import { useProjects, useSkills, useProjectSkills, useActivations, useActivate, useDeactivate, useResetProjectSkills, useProjectCategories, useDisassociateProjectCategory } from '@/infra/queries'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ConflictModal } from '@/components/conflict-modal'
import { AgentToggle } from '@/components/agent-toggle'
import { Tag, X } from 'lucide-react'
import type { Conflict, ActivateResult, Skill } from '@/types'
import { SCOPE_PROJECT, AGENT_CLAUDE, AGENT_COPILOT } from '@/types'

export const Route = createFileRoute('/projects/$id')({
  component: ProjectDetailPage,
})

function ProjectDetailPage() {
  const { id } = useParams({ from: '/projects/$id' })
  const { data: projects = [] } = useProjects()
  const { data: globalSkills = [] } = useSkills()
  const { data: projectSkills = [] } = useProjectSkills(id)
  const { data: activations = [] } = useActivations({ projectId: id })
  const activate = useActivate()
  const deactivate = useDeactivate()
  const resetSkills = useResetProjectSkills()
  const [conflict, setConflict] = useState<Conflict | null>(null)
  const [showReset, setShowReset] = useState(false)

  const project = projects.find((p) => p.id === id)
  if (!project) return <p className="text-muted-foreground">Project not found.</p>

  async function handleToggle(skillId: string, agent: string, active: boolean) {
    if (active) {
      const existing = activations.find(
        (a) => a.skillId === skillId && a.agent === agent && a.scope === SCOPE_PROJECT,
      )
      if (existing) await deactivate.mutateAsync(existing.id)
    } else {
      const result: ActivateResult = await activate.mutateAsync({
        skillId,
        agent,
        scope: SCOPE_PROJECT,
        projectId: id,
      })
      if (result.conflict) setConflict(result.conflict)
    }
  }

  function isClaudeActive(skillId: string) {
    return activations.some((a) => a.skillId === skillId && a.agent === AGENT_CLAUDE)
  }
  function isCopilotActive(skillId: string) {
    return activations.some((a) => a.skillId === skillId && a.agent === AGENT_COPILOT)
  }

  return (
    <div className="space-y-6">
      <div>
        <Link to="/projects" className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-4">
          <ArrowLeft className="h-3.5 w-3.5" /> Projects
        </Link>
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold">{project.name}</h1>
            <p className="text-xs text-muted-foreground font-mono mt-1">{project.path}</p>
            <div className="flex gap-1 mt-2">
              {project.detectedAgents.map((a) => (
                <Badge key={a} variant="secondary">{a}</Badge>
              ))}
            </div>
          </div>
          <Button
            variant="outline"
            size="sm"
            className="shrink-0 text-destructive hover:text-destructive"
            onClick={() => setShowReset(true)}
          >
            <RotateCcw className="h-3.5 w-3.5 mr-1.5" />
            Reset Skills
          </Button>
        </div>
      </div>

      <SkillSection
        title="Project Skills"
        description="Skills found inside this project's folder (skills/ and .claude/skills/)."
        skills={projectSkills}
        emptyMessage="No skills found in this project. Create a subfolder with SKILL.md inside skills/ or .claude/skills/."
        isClaudeActive={isClaudeActive}
        isCopilotActive={isCopilotActive}
        onToggle={handleToggle}
      />

      <SkillSection
        title="Global Skills"
        description="Skills available globally. Activate individually for this project."
        skills={globalSkills}
        emptyMessage="No global skills configured. Add sources in Settings."
        isClaudeActive={isClaudeActive}
        isCopilotActive={isCopilotActive}
        onToggle={handleToggle}
      />

      <LinkedCategories projectId={id} />

      {conflict && (
        <ConflictModal conflict={conflict} onClose={() => setConflict(null)} />
      )}
      {showReset && (
        <ResetModal
          projectName={project.name}
          isPending={resetSkills.isPending}
          onConfirm={async () => {
            await resetSkills.mutateAsync(id)
            setShowReset(false)
          }}
          onClose={() => setShowReset(false)}
        />
      )}
    </div>
  )
}

function ResetModal({
  projectName,
  isPending,
  onConfirm,
  onClose,
}: {
  projectName: string
  isPending: boolean
  onConfirm: () => void
  onClose: () => void
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <Card className="w-full max-w-sm mx-4">
        <CardHeader>
          <CardTitle className="text-base text-destructive">Reset project skills</CardTitle>
          <p className="text-sm text-muted-foreground">
            Removes all skills copied into <strong>{projectName}</strong> (folders{' '}
            <code>.claude/skills/</code> and <code>.github/skills/</code>) and clears all
            related activations.
          </p>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-xs text-muted-foreground">
            The original global skills are not affected.
          </p>
          <div className="flex gap-2 justify-end">
            <Button variant="outline" onClick={onClose} disabled={isPending}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={onConfirm} disabled={isPending}>
              {isPending ? 'Resetting…' : 'Confirm reset'}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

interface SkillSectionProps {
  title: string
  description: string
  skills: Skill[]
  emptyMessage: string
  isClaudeActive: (skillId: string) => boolean
  isCopilotActive: (skillId: string) => boolean
  onToggle: (skillId: string, agent: string, active: boolean) => void
}

function SkillSection({
  title,
  description,
  skills,
  emptyMessage,
  isClaudeActive,
  isCopilotActive,
  onToggle,
}: SkillSectionProps) {
  return (
    <div>
      <div className="mb-3">
        <h2 className="text-base font-medium">{title}</h2>
        <p className="text-xs text-muted-foreground mt-0.5">{description}</p>
      </div>
      {!skills.length ? (
        <p className="text-sm text-muted-foreground italic">{emptyMessage}</p>
      ) : (
        <div className="grid gap-3">
          {skills.map((skill) => (
            <div key={skill.id} className="rounded-lg border border-border px-4 py-3">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-sm font-medium">{skill.name}</p>
                  <p className="text-xs text-muted-foreground mt-0.5">{skill.description}</p>
                </div>
                <div className="flex gap-3 shrink-0">
                  <AgentToggle
                    label="Claude"
                    active={isClaudeActive(skill.id)}
                    onToggle={(v) => onToggle(skill.id, AGENT_CLAUDE, v)}
                  />
                  <AgentToggle
                    label="Copilot"
                    active={isCopilotActive(skill.id)}
                    onToggle={(v) => onToggle(skill.id, AGENT_COPILOT, v)}
                  />
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function LinkedCategories({ projectId }: { projectId: string }) {
  const { data: links = [] } = useProjectCategories(projectId)
  const disassociate = useDisassociateProjectCategory()

  if (links.length === 0) return null

  return (
    <div>
      <div className="mb-3">
        <h2 className="text-base font-medium">Linked Categories</h2>
        <p className="text-xs text-muted-foreground mt-0.5">
          Skills from these categories are automatically copied into this project.
        </p>
      </div>
      <div className="flex flex-wrap gap-2">
        {links.map((l) => (
          <div
            key={`${l.categoryId}-${l.agent}`}
            className="flex items-center gap-1.5 rounded-md border border-border px-3 py-1.5 text-sm"
          >
            <Tag className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="font-medium">{l.category.name}</span>
            <Badge variant="secondary" className="text-[10px] py-0 h-4">{l.agent}</Badge>
            <button
              className="ml-1 text-muted-foreground hover:text-destructive transition-colors"
              title="Remove link"
              onClick={() => disassociate.mutate({ projectId, categoryId: l.categoryId, agent: l.agent })}
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
