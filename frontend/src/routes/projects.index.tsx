import { createFileRoute, Link } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import { FolderOpen, Plus, Trash2, Search, RefreshCw } from 'lucide-react'
import { useProjects, useDeleteProject, useScanCandidates, useConfirmCandidate } from '@/infra/queries'
import { settingsApi } from '@/infra/settings'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

export const Route = createFileRoute('/projects/')({
  component: ProjectsPage,
})

function ProjectsPage() {
  const { data: projects = [], isLoading } = useProjects()
  const deleteProject = useDeleteProject()
  const confirmCandidate = useConfirmCandidate()

  const [configuredRoots, setConfiguredRoots] = useState<string[]>([])
  const [scanning, setScanning] = useState(false)

  useEffect(() => {
    settingsApi.get().then((s) => setConfiguredRoots(s.workspaceRoots ?? [])).catch(console.error)
  }, [])

  const { data: candidates = [] } = useScanCandidates(configuredRoots, scanning && configuredRoots.length > 0)

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
            <Search className="h-4 w-4" /> Discover Projects
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {configuredRoots.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No workspace configured. Add paths in{' '}
              <Link to="/settings" className="underline hover:text-foreground">Settings → Workspace Roots</Link>.
            </p>
          ) : (
            <>
              <div className="space-y-1">
                <p className="text-xs text-muted-foreground font-medium">Configured folders:</p>
                {configuredRoots.map((r) => (
                  <p key={r} className="text-xs font-mono text-muted-foreground pl-2">{r}</p>
                ))}
              </div>
              <Button size="sm" onClick={handleScan}>
                <RefreshCw className="h-3.5 w-3.5 mr-1" /> Scan
              </Button>
            </>
          )}

          {scanning && candidates.length > 0 && (
            <div className="mt-4 space-y-2">
              <p className="text-sm font-medium">{candidates.length} project{candidates.length !== 1 ? 's' : ''} found</p>
              {candidates.map((c) => {
                const alreadyAdded = projects.some((p) => p.path === c.path)
                return (
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
                      disabled={alreadyAdded}
                      onClick={() => confirmCandidate.mutate(c)}
                    >
                      <Plus className="h-3.5 w-3.5 mr-1" />
                      {alreadyAdded ? 'Added' : 'Add'}
                    </Button>
                  </div>
                )
              })}
            </div>
          )}
          {scanning && candidates.length === 0 && configuredRoots.length > 0 && (
            <p className="text-sm text-muted-foreground">No projects found in the configured folders.</p>
          )}
        </CardContent>
      </Card>

      {/* Registered projects */}
      {!projects.length ? (
        <div className="text-center mt-8 text-muted-foreground">
          <FolderOpen className="h-10 w-10 mx-auto mb-3 opacity-40" />
          <p className="text-sm">No projects registered yet. Use the button above to scan.</p>
        </div>
      ) : (
        <div className="grid gap-3">
          {projects.map((p) => (
            <Link key={p.id} to="/projects/$id" params={{ id: p.id }}>
              <div className="flex items-center justify-between rounded-lg border border-border px-4 py-3 hover:bg-accent transition-colors cursor-pointer">
                <div>
                  <p className="text-sm font-medium">{p.name}</p>
                  <p className="text-xs text-muted-foreground font-mono mt-0.5">{p.path}</p>
                  <div className="flex gap-1 mt-1">
                    {p.detectedAgents.map((a) => {
                      const active = p.activeAgents?.includes(a)
                      return (
                        <Badge key={a} variant={active ? 'default' : 'secondary'} className="text-xs">
                          {a}
                        </Badge>
                      )
                    })}
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  className="text-muted-foreground hover:text-destructive"
                  onClick={(e) => { e.preventDefault(); deleteProject.mutate(p.id) }}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
