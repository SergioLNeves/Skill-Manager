import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryResult,
} from '@tanstack/react-query'
import { api } from './bindings'
import type {
  Skill,
  Project,
  ProjectCandidate,
  Activation,
  ActivationFilter,
  ActivateRequest,
  ActivateResult,
  ResolveConflictRequest,
  RegisterProjectRequest,
  DoctorReport,
} from '@/types'

export const keys = {
  skills: ['skills'] as const,
  projects: ['projects'] as const,
  activations: (filter?: ActivationFilter) =>
    filter ? ['activations', filter] : (['activations'] as const),
  doctor: ['doctor'] as const,
}

// --- Skills ---

export function useSkills(): UseQueryResult<Skill[]> {
  return useQuery({ queryKey: keys.skills, queryFn: api.listSkills })
}

// --- Projects ---

export function useProjects(): UseQueryResult<Project[]> {
  return useQuery({ queryKey: keys.projects, queryFn: api.listProjects })
}

export function useScanCandidates(roots: string[], enabled: boolean) {
  return useQuery({
    queryKey: ['scan', roots],
    queryFn: () => api.scanCandidates(roots),
    enabled,
  })
}

export function useConfirmCandidate() {
  const qc = useQueryClient()
  return useMutation<Project, Error, ProjectCandidate>({
    mutationFn: api.confirmCandidate,
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.projects }),
  })
}

export function useRegisterProject() {
  const qc = useQueryClient()
  return useMutation<Project, Error, RegisterProjectRequest>({
    mutationFn: api.registerProject,
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.projects }),
  })
}

export function useDeleteProject() {
  const qc = useQueryClient()
  return useMutation<void, Error, string>({
    mutationFn: api.deleteProject,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: keys.projects })
      qc.invalidateQueries({ queryKey: keys.activations() })
    },
  })
}

// --- Activations ---

export function useActivations(filter: ActivationFilter = {}) {
  return useQuery({
    queryKey: keys.activations(filter),
    queryFn: () => api.listActivations(filter),
  })
}

export function useActivate() {
  const qc = useQueryClient()
  return useMutation<ActivateResult, Error, ActivateRequest>({
    mutationFn: api.activate,
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.activations() }),
  })
}

export function useDeactivate() {
  const qc = useQueryClient()
  return useMutation<void, Error, number>({
    mutationFn: api.deactivate,
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.activations() }),
  })
}

export function useResolveConflict() {
  const qc = useQueryClient()
  return useMutation<void, Error, ResolveConflictRequest>({
    mutationFn: api.resolveConflict,
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.activations() }),
  })
}

// --- Doctor ---

export function useDoctorReport(enabled: boolean) {
  return useQuery({
    queryKey: keys.doctor,
    queryFn: api.runDoctor,
    enabled,
  })
}

export function useRunDoctor() {
  const qc = useQueryClient()
  return useMutation<DoctorReport, Error>({
    mutationFn: api.runDoctor,
    onSuccess: (data) => qc.setQueryData(keys.doctor, data),
  })
}
