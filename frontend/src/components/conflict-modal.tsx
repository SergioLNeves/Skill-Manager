import { useResolveConflict } from '@/infra/queries'
import { Button } from '@/components/ui/button'
import type { Conflict } from '@/types'
import { RESOLUTION_KEEP_GLOBAL, RESOLUTION_OVERRIDE_PROJECT } from '@/types'

interface Props {
  conflict: Conflict
  onClose: () => void
}

export function ConflictModal({ conflict, onClose }: Props) {
  const resolve = useResolveConflict()

  async function handle(resolution: typeof RESOLUTION_KEEP_GLOBAL | typeof RESOLUTION_OVERRIDE_PROJECT) {
    await resolve.mutateAsync({ conflict, resolution })
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-xl border border-border bg-background shadow-xl p-6">
        <h2 className="text-base font-semibold mb-1">Activation Conflict</h2>
        <p className="text-sm text-muted-foreground mb-4">
          <strong>{conflict.skillId}</strong> is already active globally for{' '}
          <strong>{conflict.agent}</strong>. How do you want to resolve this?
        </p>

        <div className="space-y-2 mb-6 text-sm">
          {conflict.globalActivation && (
            <div className="rounded-md border border-border px-3 py-2">
              <span className="text-xs uppercase tracking-wide text-muted-foreground">Global</span>
              <p className="mt-0.5">Active for all projects</p>
            </div>
          )}
          {conflict.projectActivation && (
            <div className="rounded-md border border-border px-3 py-2">
              <span className="text-xs uppercase tracking-wide text-muted-foreground">Project</span>
              <p className="mt-0.5">Project: {conflict.projectActivation.projectId}</p>
            </div>
          )}
        </div>

        <div className="flex gap-2 justify-end">
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="outline"
            onClick={() => handle(RESOLUTION_KEEP_GLOBAL)}
            disabled={resolve.isPending}
          >
            Keep Global
          </Button>
          <Button
            onClick={() => handle(RESOLUTION_OVERRIDE_PROJECT)}
            disabled={resolve.isPending}
          >
            Override in Project
          </Button>
        </div>
      </div>
    </div>
  )
}
