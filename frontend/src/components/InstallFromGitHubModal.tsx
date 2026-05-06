import { useState } from 'react'
import { GitFork, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useInstallGitHubSkill } from '@/infra/queries'

interface Props {
  onClose: () => void
}

export function InstallFromGitHubModal({ onClose }: Props) {
  const [input, setInput] = useState('')
  const install = useInstallGitHubSkill()

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!input.trim()) return
    install.mutate(input.trim(), {
      onSuccess: (names) => {
        setInput('')
        onClose()
        // Brief feedback visible via toast if needed; for now just close.
        console.info('Installed skills:', names)
      },
    })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-background border border-border rounded-lg shadow-lg w-full max-w-md p-6 space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <GitFork className="h-5 w-5" />
            <h2 className="text-base font-semibold">Instalar skill do GitHub</h2>
          </div>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>

        <p className="text-sm text-muted-foreground">
          Cole o repositório ou o comando completo. A skill será instalada globalmente.
        </p>

        <form onSubmit={handleSubmit} className="space-y-3">
          <input
            autoFocus
            type="text"
            placeholder="ex: owner/repo ou npx skills add owner/repo --ref main"
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            disabled={install.isPending}
          />

          {install.isError && (
            <p className="text-sm text-destructive">{install.error.message}</p>
          )}

          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={onClose} disabled={install.isPending}>
              Cancelar
            </Button>
            <Button type="submit" disabled={!input.trim() || install.isPending}>
              {install.isPending ? 'Instalando…' : 'Instalar'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
