import type { Settings } from '@/types'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const go = (window as any).go?.main?.App

function call<T>(method: string, ...args: unknown[]): Promise<T> {
  if (!go) throw new Error('Wails runtime not available')
  return go[method](...args) as Promise<T>
}

export const settingsApi = {
  get: () => call<Settings>('GetSettings'),
  save: (s: Settings) => call<void>('SaveSettings', s),
}
