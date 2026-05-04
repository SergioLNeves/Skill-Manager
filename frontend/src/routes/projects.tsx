import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { FolderOpen, Plus, Trash2, Search } from 'lucide-react'
import { useProjects, useDeleteProject, useScanCandidates, useConfirmCandidate } from '@/infra/queries'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

export const Route = createFileRoute('/projects')({
  component: ProjectsPage,
})

function ProjectsPage() {
  const { data: projects = [], isLoading } = useProjects()
  const deleteProject = useDeleteProject()
  const confirmCandidate = useConfirmCandidate()

  const [scanRoots, setScanRoots] = useState('')
  const [scanning, setScanning] = useState(false)
  const roots = scanRoots.split('\n').map((r) => r.trim()).filter(Boolean)
  const { data: candidates = [] } = useScanCandidates(roots, scanning && roots.length > 0)

  function handleScan() {
    setScanning(true)
  }

  if (isLoading) return <p className="text-muted-foreground">Loading projects…</p>

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Projects</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {projects.length} registered project{projects.length !== 1 ? 's' : ''}
          </p>
        </div>
      </div>

      {/* Scan workspace */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Search className="h-4 w-4" /> Scan Workspace
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <textarea
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono resize-none h-20 focus:outline-none focus:ring-2 focus:ring-ring"
            placeholder="One workspace root per line, e.g. /home/user/dev"
            value={scanRoots}
            onChange={(e) => setScanRoots(e.target.value)}
          />
          <Button size="sm" onClick={handleScan} disabled={!roots.length}>
            <Search className="h-3.5 w-3.5 mr-1" /> Scan
          </Button>

          {scanning && candidates.length > 0 && (
            <div className="mt-4 space-y-2">
              <p className="text-sm font-medium">{candidates.length} candidates found</p>
              {candidates.map((c) => (
                <div key={c.path} className="flex items-center justify-between rounded-md border border-border px-3 py-2">
                  <div>
                    <p className="text-sm font-medium">{c.name}</p>
                    <p className="text-xs text-muted-foreground font-mono">{c.path}</p>
                    <div className="flex gap-1 mt-1">
                      {c.detectedAgents.map((a) => (
                        <Badge key={a} variant="secondary" className="text-xs">{a}</Badge>
                      ))}
                    </div>
                  </div>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => confirmCandidate.mutate(c)}
                  >
                    <Plus className="h-3.5 w-3.5 mr-1" /> Add
                  </Button>
                </div>
              ))}
            </div>
          )}
          {scanning && candidates.length === 0 && (
            <p className="text-sm text-muted-foreground">No projects found in the given roots.</p>
          )}
        </CardContent>
      </Card>

      {/* Registered projects */}
      {!projects.length ? (
        <div className="text-center mt-8 text-muted-foreground">
          <FolderOpen className="h-10 w-10 mx-auto mb-3 opacity-40" />
          <p className="text-sm">No projects registered yet. Scan a workspace to get started.</p>
        </div>
      ) : (
        <div className="grid gap-3">
          {projects.map((p) => (
            <div key={p.id} className="flex items-center justify-between rounded-lg border border-border px-4 py-3">
              <div>
                <Link to="/projects/$id" params={{ id: p.id }} className="text-sm font-medium hover:underline">
                  {p.name}
                </Link>
                <p className="text-xs text-muted-foreground font-mono mt-0.5">{p.path}</p>
                <div className="flex gap-1 mt-1">
                  {p.detectedAgents.map((a) => (
                    <Badge key={a} variant="secondary" className="text-xs">{a}</Badge>
                  ))}
                </div>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="text-muted-foreground hover:text-destructive"
                onClick={() => deleteProject.mutate(p.id)}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
