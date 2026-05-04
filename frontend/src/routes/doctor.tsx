import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { ShieldCheck, CheckCircle2, AlertTriangle, XCircle, Wrench, ChevronDown, ChevronRight, RefreshCw } from 'lucide-react'
import { useRunDoctor, useFixIssue } from '@/infra/queries'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import type { DoctorIssue } from '@/types'

export const Route = createFileRoute('/doctor')({
  component: DoctorPage,
})

const KIND_META: Record<string, { label: string; severity: 'warning' | 'error' }> = {
  orphaned_activation: { label: 'Ativação órfã', severity: 'warning' },
  missing_project:     { label: 'Projeto ausente', severity: 'warning' },
  broken_symlink:      { label: 'Symlink quebrado', severity: 'warning' },
  missing_project_path:{ label: 'Diretório não encontrado', severity: 'error' },
}

function DoctorPage() {
  const runDoctor = useRunDoctor()
  const fixIssue = useFixIssue()
  const report = runDoctor.data

  const fixingAll = fixIssue.isPending

  async function handleFixAll() {
    if (!report) return
    for (const issue of report.issues.filter((i) => i.fixable)) {
      await fixIssue.mutateAsync(issue)
    }
    runDoctor.mutate()
  }

  const fixableCount = report?.issues.filter((i) => i.fixable).length ?? 0

  return (
    <div className="space-y-6 max-w-2xl">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">Doctor</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Verifica a consistência entre o banco de dados e o sistema de arquivos.
          </p>
        </div>
        <Button onClick={() => runDoctor.mutate()} disabled={runDoctor.isPending}>
          <ShieldCheck className="h-4 w-4 mr-2" />
          {runDoctor.isPending ? 'Verificando…' : 'Verificar agora'}
        </Button>
      </div>

      {/* Estado inicial */}
      {!report && !runDoctor.isPending && (
        <div className="rounded-lg border border-dashed border-border p-12 text-center text-muted-foreground">
          <ShieldCheck className="h-10 w-10 mx-auto mb-3 opacity-30" />
          <p className="text-sm">Clique em "Verificar agora" para diagnosticar possíveis problemas.</p>
        </div>
      )}

      {/* Tudo OK */}
      {report && report.issues.length === 0 && (
        <div className="rounded-lg border border-border p-10 text-center space-y-2">
          <CheckCircle2 className="h-9 w-9 mx-auto text-green-500" />
          <p className="text-sm font-medium">Tudo certo! Nenhum problema encontrado.</p>
          <p className="text-xs text-muted-foreground">O registro está consistente com o sistema de arquivos.</p>
        </div>
      )}

      {/* Issues encontradas */}
      {report && report.issues.length > 0 && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">
              <span className="font-medium text-foreground">{report.issues.length}</span>{' '}
              problema{report.issues.length !== 1 ? 's' : ''} encontrado{report.issues.length !== 1 ? 's' : ''}
            </p>
            {fixableCount > 0 && (
              <Button
                size="sm"
                variant="outline"
                onClick={handleFixAll}
                disabled={fixingAll}
              >
                <Wrench className="h-3.5 w-3.5 mr-1.5" />
                {fixingAll ? 'Corrigindo…' : `Corrigir tudo (${fixableCount})`}
              </Button>
            )}
          </div>

          <div className="space-y-3">
            {report.issues.map((issue, i) => (
              <IssueCard key={i} issue={issue} />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function IssueCard({ issue }: { issue: DoctorIssue }) {
  const [expanded, setExpanded] = useState(false)
  const fixIssue = useFixIssue()
  const [fixed, setFixed] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const meta = KIND_META[issue.kind] ?? { label: issue.kind, severity: 'warning' }
  const isWarning = meta.severity === 'warning'

  async function handleFix() {
    try {
      await fixIssue.mutateAsync(issue)
      setFixed(true)
    } catch (e) {
      setError(String(e))
    }
  }

  if (fixed) {
    return (
      <div className="flex items-center gap-3 rounded-lg border border-border px-4 py-3 bg-muted/30">
        <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0" />
        <p className="text-sm text-muted-foreground">{issue.title} — corrigido.</p>
      </div>
    )
  }

  return (
    <div className={cn(
      'rounded-lg border overflow-hidden',
      isWarning ? 'border-yellow-500/30' : 'border-destructive/30',
    )}>
      {/* Header sempre visível */}
      <div className="flex items-start gap-3 px-4 py-3">
        {isWarning
          ? <AlertTriangle className="h-4 w-4 mt-0.5 shrink-0 text-yellow-500" />
          : <XCircle className="h-4 w-4 mt-0.5 shrink-0 text-destructive" />
        }

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <p className="text-sm font-medium">{issue.title}</p>
            <Badge
              variant="outline"
              className={cn('text-[10px] py-0 h-4', isWarning ? 'border-yellow-500/50 text-yellow-600' : 'border-destructive/50 text-destructive')}
            >
              {meta.label}
            </Badge>
            {issue.fixable && (
              <Badge variant="secondary" className="text-[10px] py-0 h-4">auto-fix disponível</Badge>
            )}
          </div>
          <p className="text-xs text-muted-foreground mt-0.5">{issue.detail}</p>
        </div>

        <button
          onClick={() => setExpanded((v) => !v)}
          className="shrink-0 text-muted-foreground hover:text-foreground mt-0.5"
        >
          {expanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
        </button>
      </div>

      {/* Detalhes expandidos */}
      {expanded && (
        <div className={cn(
          'border-t px-4 py-3 space-y-3',
          isWarning ? 'border-yellow-500/20 bg-yellow-50/20' : 'border-destructive/20 bg-destructive/5',
        )}>
          {issue.fixable && issue.howToFix && (
            <div className="space-y-1">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Como corrigir</p>
              <p className="text-sm">{issue.howToFix}</p>
            </div>
          )}

          {!issue.fixable && (
            <div className="space-y-1">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">Correção manual necessária</p>
              <p className="text-sm text-muted-foreground">
                Este problema não pode ser corrigido automaticamente. Verifique o diretório do projeto e remova-o do registro se não for mais necessário.
              </p>
            </div>
          )}

          {error && <p className="text-xs text-destructive">{error}</p>}

          <div className="flex gap-2">
            {issue.fixable && (
              <Button
                size="sm"
                variant={isWarning ? 'outline' : 'destructive'}
                onClick={handleFix}
                disabled={fixIssue.isPending}
              >
                <Wrench className="h-3.5 w-3.5 mr-1.5" />
                {fixIssue.isPending ? 'Corrigindo…' : 'Corrigir este problema'}
              </Button>
            )}
            <Button size="sm" variant="ghost" className="text-muted-foreground" onClick={() => setExpanded(false)}>
              Fechar
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
