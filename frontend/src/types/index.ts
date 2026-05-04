export interface Skill {
  id: string
  name: string
  description: string
  path: string
  updatedAt: string
}

export interface Project {
  id: string
  name: string
  path: string
  detectedAgents: string[]
  addedAt: string
}

export interface ProjectCandidate {
  name: string
  path: string
  detectedAgents: string[]
}

export interface Activation {
  id: number
  skillId: string
  agent: string
  scope: string
  projectId: string
  appliedAt: string
}

export interface ActivationFilter {
  skillId?: string
  agent?: string
  scope?: string
  projectId?: string
}

export interface ActivateRequest {
  skillId: string
  agent: string
  scope: string
  projectId: string
}

export interface ConflictActivation {
  id: number
  skillId: string
  agent: string
  scope: string
  projectId: string
  appliedAt: string
}

export interface Conflict {
  skillId: string
  agent: string
  globalActivation: ConflictActivation | null
  projectActivation: ConflictActivation | null
}

export interface ActivateResult {
  activation: Activation | null
  conflict: Conflict | null
}

export interface ResolveConflictRequest {
  conflict: Conflict
  resolution: 1 | 2 // 1=KeepGlobal, 2=OverrideProject
}

export interface RegisterProjectRequest {
  path: string
  detectedAgents: string[]
}

export interface DoctorIssue {
  kind: string
  detail: string
  fixable: boolean
}

export interface DoctorReport {
  issues: DoctorIssue[]
}

export interface Settings {
  workspaceRoots: string[]
  skillsHome: string
}

export const AGENT_CLAUDE = 'claude'
export const AGENT_COPILOT = 'copilot'

export const SCOPE_GLOBAL = 'global'
export const SCOPE_PROJECT = 'project'

export const RESOLUTION_KEEP_GLOBAL = 1 as const
export const RESOLUTION_OVERRIDE_PROJECT = 2 as const
