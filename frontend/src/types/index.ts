export interface Skill {
  id: string
  name: string
  description: string
  categoryId: number | null
  categoryName: string
  path: string
  source: 'global' | 'project'
  ownerProjectId: string
  ownerProjectName: string
  updatedAt: string
}

export interface SkillProjectRef {
  id: string
  name: string
  path: string      // project root path
  skillPath: string // skill directory path (for reading SKILL.md)
}

export interface AggregatedSkill {
  name: string
  description: string
  categoryId: number | null
  categoryName: string
  isGlobal: boolean
  globalPath: string
  projects: SkillProjectRef[]
  updatedAt: string
}

export interface Category {
  id: number
  name: string
  description: string
  createdAt: string
}

export interface ProjectCategoryLink {
  projectId: string
  categoryId: number
  agent: string
  category: Category
}

export interface CreateCategoryRequest {
  name: string
  description: string
}

export interface UpdateCategoryRequest {
  id: number
  name: string
  description: string
}

export interface AssignSkillCategoryRequest {
  skillName: string
  skillPath: string
  categoryId: number | null
}

export interface ProjectCategoryRequest {
  projectId: string
  categoryId: number
  agent: string
}

export interface CopySkillRequest {
  skillId: string
  sourceProjectId: string
  targetProjectId: string
  agent: 'claude' | 'copilot'
}

export interface DeleteSkillRequest {
  skillId: string
  projectId: string
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
  title: string
  detail: string
  howToFix: string
  fixable: boolean
  fixData: Record<string, string>
}

export interface DoctorReport {
  issues: DoctorIssue[]
}

export interface Settings {
  workspaceRoots: string[]
  globalSkillSources: string[]
  // Legacy fields (back-compat)
  skillsHome?: string
  skillSources?: string[]
}

export const AGENT_CLAUDE = 'claude'
export const AGENT_COPILOT = 'copilot'

export const SCOPE_GLOBAL = 'global'
export const SCOPE_PROJECT = 'project'

export const RESOLUTION_KEEP_GLOBAL = 1 as const
export const RESOLUTION_OVERRIDE_PROJECT = 2 as const
