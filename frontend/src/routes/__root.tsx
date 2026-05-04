import { createRootRoute, Link, Outlet } from '@tanstack/react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cn } from '@/lib/utils'

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, staleTime: 30_000 } },
})

const navLinks = [
  { to: '/', label: 'Skills' },
  { to: '/projects', label: 'Projects' },
  { to: '/doctor', label: 'Doctor' },
  { to: '/settings', label: 'Settings' },
]

export const Route = createRootRoute({
  component: Root,
})

function Root() {
  return (
    <QueryClientProvider client={queryClient}>
      <div className="flex h-screen bg-background text-foreground">
        <aside className="w-48 shrink-0 border-r border-border flex flex-col gap-1 px-3 py-6">
          <span className="px-3 mb-4 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
            Skills Manager
          </span>
          {navLinks.map(({ to, label }) => (
            <Link
              key={to}
              to={to}
              className={cn(
                'rounded-md px-3 py-2 text-sm font-medium transition-colors',
                'hover:bg-accent hover:text-accent-foreground',
                '[&.active]:bg-accent [&.active]:text-accent-foreground',
              )}
            >
              {label}
            </Link>
          ))}
        </aside>
        <main className="flex-1 overflow-y-auto p-8">
          <Outlet />
        </main>
      </div>
    </QueryClientProvider>
  )
}
