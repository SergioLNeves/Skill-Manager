import { createFileRoute, useParams } from '@tanstack/react-router'
import { useState } from 'react'
import { ArrowLeft } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import { useProjects, useSkills, useActivations, useActivate, useDeactivate } from '@/infra/queries'
import { Badge } from '@/components/ui/badge'
import { ConflictModal } from '@/components/conflict-modal'
import { AgentToggle } from '@/components/agent-toggle'
import type { Conflict, ActivateResult } from '@/types'
import { SCOPE_PROJECT, AGENT_CLAUDE, AGENT_COPILOT } from '@/types'

export const Route = createFileRoute('/projects/$id')({
  component: ProjectDetailPage,
})

function ProjectDetailPage() {
  const { id } = useParams({ from: '/projects/$id' })
  const { data: projects = [] } = useProjects()
  const { data: skills = [] } = useSkills()
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

  const supportsClaude = project.detectedAgents.includes(AGENT_CLAUDE)
  const supportsCopilot = project.detectedAgents.includes(AGENT_COPILOT)

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

      <div>
        <h2 className="text-base font-medium mb-3">Skills</h2>
        {!skills.length && (
          <p className="text-sm text-muted-foreground">No skills available.</p>
        )}
        <div className="grid gap-3">
          {skills.map((skill) => {
            const claudeActive = activations.some(
              (a) => a.skillId === skill.id && a.agent === AGENT_CLAUDE,
            )
            const copilotActive = activations.some(
              (a) => a.skillId === skill.id && a.agent === AGENT_COPILOT,
            )
            return (
              <div key={skill.id} className="rounded-lg border border-border px-4 py-3">
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <p className="text-sm font-medium">{skill.name}</p>
                    <p className="text-xs text-muted-foreground mt-0.5">{skill.description}</p>
                  </div>
                  <div className="flex gap-3 shrink-0">
                    <AgentToggle
                      label="Claude"
                      active={claudeActive}
                      disabled={!supportsClaude}
                      onToggle={(v) => handleToggle(skill.id, AGENT_CLAUDE, v)}
                    />
                    <AgentToggle
                      label="Copilot"
                      active={copilotActive}
                      disabled={!supportsCopilot}
                      onToggle={(v) => handleToggle(skill.id, AGENT_COPILOT, v)}
                    />
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>

      {conflict && (
        <ConflictModal conflict={conflict} onClose={() => setConflict(null)} />
      )}
    </div>
  )
}
