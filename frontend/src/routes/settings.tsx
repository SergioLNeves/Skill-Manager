import { createFileRoute } from '@tanstack/react-router'
import { useState, useEffect } from 'react'
import { Save } from 'lucide-react'
import { settingsApi } from '@/infra/settings'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { Settings } from '@/types'

export const Route = createFileRoute('/settings')({
  component: SettingsPage,
})

function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null)
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    settingsApi.get().then(setSettings).catch(console.error)
  }, [])

  async function handleSave() {
    if (!settings) return
    setSaving(true)
    try {
      await settingsApi.save(settings)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } finally {
      setSaving(false)
    }
  }

  if (!settings) return <p className="text-muted-foreground">Loading settings…</p>

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-semibold">Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Application configuration. Restart the app to apply changes to paths.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Skills Repository</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <label className="text-sm font-medium">Skills Home Path</label>
          <input
            type="text"
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-ring"
            value={settings.skillsHome}
            onChange={(e) => setSettings({ ...settings, skillsHome: e.target.value })}
          />
          <p className="text-xs text-muted-foreground">
            Override with <code>SKILLS_MANAGER_HOME</code> environment variable.
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Workspace Roots</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          <label className="text-sm font-medium">Default roots for project scanning</label>
          <textarea
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono resize-none h-24 focus:outline-none focus:ring-2 focus:ring-ring"
            placeholder="One path per line"
            value={settings.workspaceRoots.join('\n')}
            onChange={(e) =>
              setSettings({
                ...settings,
                workspaceRoots: e.target.value.split('\n').map((r) => r.trim()).filter(Boolean),
              })
            }
          />
        </CardContent>
      </Card>

      <div className="flex items-center gap-3">
        <Button onClick={handleSave} disabled={saving}>
          <Save className="h-4 w-4 mr-2" />
          {saving ? 'Saving…' : 'Save Settings'}
        </Button>
        {saved && <span className="text-sm text-green-600">Saved!</span>}
      </div>
    </div>
  )
}
