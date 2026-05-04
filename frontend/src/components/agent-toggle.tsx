import { cn } from '@/lib/utils'

interface Props {
  label: string
  active: boolean
  disabled?: boolean
  onToggle: (active: boolean) => void
}

export function AgentToggle({ label, active, disabled, onToggle }: Props) {
  return (
    <button
      type="button"
      onClick={() => !disabled && onToggle(active)}
      disabled={disabled}
      className={cn(
        'flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium transition-colors border',
        active
          ? 'bg-primary text-primary-foreground border-primary'
          : 'bg-transparent text-muted-foreground border-border hover:border-primary/50',
        disabled && 'opacity-30 cursor-not-allowed',
      )}
      title={disabled ? `${label} not detected in this project` : undefined}
    >
      <span
        className={cn(
          'h-1.5 w-1.5 rounded-full',
          active ? 'bg-primary-foreground' : 'bg-muted-foreground',
        )}
      />
      {label}
    </button>
  )
}
