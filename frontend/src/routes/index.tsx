import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { useSkills } from '@/infra/queries'
import { useActivations, useActivate, useDeactivate } from '@/infra/queries'
import { SkillCard } from '@/components/skill-card'
import { ConflictModal } from '@/components/conflict-modal'
import type { Conflict, ActivateResult } from '@/types'
import { AGENT_CLAUDE, AGENT_COPILOT, SCOPE_GLOBAL } from '@/types'

export const Route = createFileRoute('/')({
  component: SkillsPage,
})

function SkillsPage() {
  const { data: skills = [], isLoading } = useSkills()
  const { data: activations = [] } = useActivations()
  const activate = useActivate()
  const deactivate = useDeactivate()
  const [conflict, setConflict] = useState<Conflict | null>(null)

  const activeIds = new Set(activations.map((a) => a.skillId))

  async function handleToggle(skillId: string, agent: string, active: boolean) {
    if (active) {
      const existing = activations.find(
        (a) => a.skillId === skillId && a.agent === agent,
      )
      if (existing) {
        await deactivate.mutateAsync(existing.id)
      }
    } else {
      const result: ActivateResult = await activate.mutateAsync({
        skillId,
        agent,
        scope: SCOPE_GLOBAL,
        projectId: '',
      })
      if (result.conflict) {
        setConflict(result.conflict)
      }
    }
  }

  if (isLoading) return <p className="text-muted-foreground">Loading skills…</p>
  if (!skills.length)
    return (
      <div className="text-center mt-16 text-muted-foreground">
        <p className="text-lg font-medium">No skills found</p>
        <p className="text-sm mt-1">
          Add skills to <code>~/.skills-manager/skills/</code>
        </p>
      </div>
    )

  return (
    <>
      <div className="mb-6">
        <h1 className="text-2xl font-semibold">Skills</h1>
        <p className="text-sm text-muted-foreground mt-1">
          {skills.length} skill{skills.length !== 1 ? 's' : ''} available
        </p>
      </div>

      <div className="grid gap-4">
        {skills.map((skill) => {
          const claudeActive = activations.some(
            (a) =>
              a.skillId === skill.id &&
              a.agent === AGENT_CLAUDE &&
              a.scope === SCOPE_GLOBAL,
          )
          const copilotActive = activations.some(
            (a) =>
              a.skillId === skill.id &&
              a.agent === AGENT_COPILOT &&
              a.scope === SCOPE_GLOBAL,
          )
          return (
            <SkillCard
              key={skill.id}
              skill={skill}
              claudeActive={claudeActive}
              copilotActive={copilotActive}
              onToggleClaude={(active) =>
                handleToggle(skill.id, AGENT_CLAUDE, active)
              }
              onToggleCopilot={(active) =>
                handleToggle(skill.id, AGENT_COPILOT, active)
              }
            />
          )
        })}
      </div>

      {conflict && (
        <ConflictModal
          conflict={conflict}
          onClose={() => setConflict(null)}
        />
      )}
    </>
  )
}
