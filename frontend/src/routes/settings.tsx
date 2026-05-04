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
      // SelectDirectory pode falhar em alguns ambientes de dev — usuário digita manualmente
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
                title="Selecionar pasta"
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
                title="Remover"
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
            Selecionar pasta
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onChange([...paths, ''])}
          >
            <Plus className="h-4 w-4 mr-2" />
            Digitar caminho
          </Button>
        </div>
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
  if (!settings) return <p className="text-muted-foreground">Carregando configurações…</p>

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-semibold">Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Reinicie o app para aplicar alterações de caminhos.
        </p>
      </div>

      <PathList
        label="Skills Globais"
        description="Pastas varridas recursivamente para descobrir skills disponíveis para todos os projetos. Qualquer subpasta com SKILL.md é reconhecida como uma skill."
        paths={settings.globalSkillSources}
        onChange={(v) => setSettings({ ...settings, globalSkillSources: v })}
      />

      <PathList
        label="Workspace Roots"
        description="Raízes do ambiente de desenvolvimento. O app escaneia essas pastas para descobrir projetos (detecta repositórios .git)."
        paths={settings.workspaceRoots}
        onChange={(v) => setSettings({ ...settings, workspaceRoots: v })}
      />

      <div className="flex items-center gap-3">
        <Button onClick={handleSave} disabled={saving}>
          <Save className="h-4 w-4 mr-2" />
          {saving ? 'Salvando…' : 'Salvar'}
        </Button>
        {saved && <span className="text-sm text-green-600">Salvo!</span>}
      </div>
    </div>
  )
}
