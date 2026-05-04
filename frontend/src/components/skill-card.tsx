import type { Skill } from '@/types'
import { AgentToggle } from './agent-toggle'

interface Props {
  skill: Skill
  claudeActive: boolean
  copilotActive: boolean
  onToggleClaude: (active: boolean) => void
  onToggleCopilot: (active: boolean) => void
}

export function SkillCard({ skill, claudeActive, copilotActive, onToggleClaude, onToggleCopilot }: Props) {
  return (
    <div className="rounded-lg border border-border bg-card px-5 py-4">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="text-sm font-medium truncate">{skill.name}</p>
          <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">{skill.description}</p>
        </div>
        <div className="flex gap-3 shrink-0">
          <AgentToggle label="Claude" active={claudeActive} onToggle={onToggleClaude} />
          <AgentToggle label="Copilot" active={copilotActive} onToggle={onToggleCopilot} />
        </div>
      </div>
    </div>
  )
}
