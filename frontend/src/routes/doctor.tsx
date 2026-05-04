import { createFileRoute } from '@tanstack/react-router'
import { ShieldCheck, AlertTriangle, CheckCircle2 } from 'lucide-react'
import { useRunDoctor } from '@/infra/queries'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/doctor')({
  component: DoctorPage,
})

function DoctorPage() {
  const runDoctor = useRunDoctor()
  const report = runDoctor.data

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Doctor</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Check consistency between the registry and the filesystem.
          </p>
        </div>
        <Button onClick={() => runDoctor.mutate()} disabled={runDoctor.isPending}>
          <ShieldCheck className="h-4 w-4 mr-2" />
          {runDoctor.isPending ? 'Running…' : 'Run Check'}
        </Button>
      </div>

      {!report && !runDoctor.isPending && (
        <div className="rounded-lg border border-dashed border-border p-12 text-center text-muted-foreground">
          <ShieldCheck className="h-10 w-10 mx-auto mb-3 opacity-40" />
          <p className="text-sm">Run a check to see if everything is in order.</p>
        </div>
      )}

      {report && (
        <>
          {report.issues.length === 0 ? (
            <div className="rounded-lg border border-border p-8 text-center">
              <CheckCircle2 className="h-8 w-8 mx-auto mb-2 text-green-500" />
              <p className="text-sm font-medium">All good! No issues found.</p>
            </div>
          ) : (
            <div className="space-y-2">
              <p className="text-sm font-medium text-muted-foreground">
                {report.issues.length} issue{report.issues.length !== 1 ? 's' : ''} found
              </p>
              {report.issues.map((issue, i) => (
                <div
                  key={i}
                  className={cn(
                    'flex items-start gap-3 rounded-lg border px-4 py-3',
                    issue.fixable ? 'border-yellow-300/50 bg-yellow-50/30' : 'border-red-300/50 bg-red-50/30',
                  )}
                >
                  <AlertTriangle
                    className={cn(
                      'h-4 w-4 mt-0.5 shrink-0',
                      issue.fixable ? 'text-yellow-500' : 'text-red-500',
                    )}
                  />
                  <div>
                    <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                      {issue.kind.replace(/_/g, ' ')}
                    </p>
                    <p className="text-sm mt-0.5">{issue.detail}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}
