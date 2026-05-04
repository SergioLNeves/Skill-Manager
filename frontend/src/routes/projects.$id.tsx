import { createFileRoute, useParams } from '@tanstack/react-router'
import { useState } from 'react'
import { ArrowLeft } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import { useProjects, useSkills, useProjectSkills, useActivations, useActivate, useDeactivate } from '@/infra/queries'
import { Badge } from '@/components/ui/badge'
import { ConflictModal } from '@/components/conflict-modal'
import { AgentToggle } from '@/components/agent-toggle'
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
  const [conflict, setConflict] = useState<Conflict | null>(null)

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
        <h1 className="text-2xl font-semibold">{project.name}</h1>
        <p className="text-xs text-muted-foreground font-mono mt-1">{project.path}</p>
        <div className="flex gap-1 mt-2">
          {project.detectedAgents.map((a) => (
            <Badge key={a} variant="secondary">{a}</Badge>
          ))}
        </div>
      </div>

      <SkillSection
        title="Skills deste Projeto"
        description="Skills encontradas dentro da pasta do projeto (skills/ e .claude/skills/)."
        skills={projectSkills}
        emptyMessage="Nenhuma skill encontrada neste projeto. Crie uma subpasta com SKILL.md em skills/ ou .claude/skills/."
        isClaudeActive={isClaudeActive}
        isCopilotActive={isCopilotActive}
        onToggle={handleToggle}
      />

      <SkillSection
        title="Skills Globais"
        description="Skills disponíveis globalmente. Ative individualmente para este projeto."
        skills={globalSkills}
        emptyMessage="Nenhuma skill global configurada. Adicione fontes em Settings."
        isClaudeActive={isClaudeActive}
        isCopilotActive={isCopilotActive}
        onToggle={handleToggle}
      />

      {conflict && (
        <ConflictModal conflict={conflict} onClose={() => setConflict(null)} />
      )}
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
