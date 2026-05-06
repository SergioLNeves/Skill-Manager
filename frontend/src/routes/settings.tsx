import { createFileRoute } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import { Save, FolderOpen, Plus, X } from 'lucide-react'
import { settingsApi } from '@/infra/settings'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { SelectDirectory } from '../../wailsjs/go/main/App'
import type { Settings } from '@/types'

export const Route = createFileRoute('/settings')({
  component: SettingsPage,
})

function normalizeSettings(raw: Settings): Settings {
  return {
    ...raw,
    workspaceRoots: raw.workspaceRoots ?? [],
    globalSkillSources: raw.globalSkillSources ?? [],
  }
}

function PathList({
  label,
  description,
  paths,
  onChange,
}: {
  label: string
  description: string
  paths: string[]
  onChange: (paths: string[]) => void
}) {
  async function addViaDialog() {
    try {
      const path = await SelectDirectory()
      if (path) onChange([...paths, path])
    } catch {
      // SelectDirectory may fail in some dev environments — user can type the path manually
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{label}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-sm text-muted-foreground">{description}</p>
        <div className="space-y-2">
          {paths.map((p, i) => (
            <div key={i} className="flex gap-2">
              <input
                type="text"
                className="flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-ring"
                value={p}
                onChange={(e) => {
                  const next = [...paths]
                  next[i] = e.target.value
                  onChange(next)
                }}
              />
              <Button
                variant="outline"
                size="icon"
                title="Select folder"
                onClick={async () => {
                  try {
                    const picked = await SelectDirectory()
                    if (picked) {
                      const next = [...paths]
                      next[i] = picked
                      onChange(next)
                    }
                  } catch { /* silently ignore */ }
                }}
              >
                <FolderOpen className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                size="icon"
                title="Remove"
                onClick={() => onChange(paths.filter((_, j) => j !== i))}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={addViaDialog}>
            <FolderOpen className="h-4 w-4 mr-2" />
            Select folder
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onChange([...paths, ''])}
          >
            <Plus className="h-4 w-4 mr-2" />
            Enter path
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

function SkillSourcesInfo() {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Skill Sources (GitHub)</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-sm text-muted-foreground">
          Skills are installed per-project from GitHub repositories using the CLI.
          Run the command below inside a project directory to install skills:
        </p>
        <pre className="rounded-md bg-muted px-4 py-3 text-sm font-mono">
          {`skill-manager skills add <owner/repo> [--skill <name>] [--ref <branch>]`}
        </pre>
        <p className="text-sm text-muted-foreground">
          Installed skills are stored in <code className="font-mono text-xs">.skills/cache/</code> and
          tracked in <code className="font-mono text-xs">skills-lock.json</code> in the project directory.
          Use <code className="font-mono text-xs">skill-manager skills sync</code> to restore them on another machine.
        </p>
      </CardContent>
    </Card>
  )
}

function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null)
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    settingsApi
      .get()
      .then((s) => setSettings(normalizeSettings(s)))
      .catch((e) => setError(String(e)))
  }, [])

  async function handleSave() {
    if (!settings) return
    setSaving(true)
    setError(null)
    try {
      await settingsApi.save(settings)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch (e) {
      setError(String(e))
    } finally {
      setSaving(false)
    }
  }

  if (error) return <p className="text-destructive text-sm">{error}</p>
  if (!settings) return <p className="text-muted-foreground">Loading settings…</p>

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-semibold">Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Restart the app to apply path changes.
        </p>
      </div>

      <SkillSourcesInfo />

      <PathList
        label="Workspace Roots"
        description="Root folders of your development workspace. The app scans these directories to discover projects (detects .git repositories)."
        paths={settings.workspaceRoots}
        onChange={(v) => setSettings({ ...settings, workspaceRoots: v })}
      />

      <div className="flex items-center gap-3">
        <Button onClick={handleSave} disabled={saving}>
          <Save className="h-4 w-4 mr-2" />
          {saving ? 'Saving…' : 'Save'}
        </Button>
        {saved && <span className="text-sm text-green-600">Saved</span>}
      </div>
    </div>
  )
}
