import { FolderGit2 } from 'lucide-react'
import type { Host } from '@/lib/api'
import { cn } from '@/lib/utils'

export function HostIcon({ host, className }: { host: Host; className?: string }) {
  if (host === 'github') {
    return (
      <svg viewBox="0 0 16 16" aria-hidden="true" className={cn('fill-current', className)}>
        <path d="M8 0a8 8 0 0 0-2.53 15.59c.4.07.55-.17.55-.38v-1.33c-2.23.48-2.7-1.07-2.7-1.07-.36-.92-.89-1.17-.89-1.17-.73-.5.05-.49.05-.49.81.06 1.23.83 1.23.83.72 1.23 1.88.87 2.34.67.07-.52.28-.87.51-1.07-1.78-.2-3.65-.89-3.65-3.96 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82a7.6 7.6 0 0 1 4 0c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.08-1.87 3.75-3.66 3.95.29.25.54.73.54 1.48v2.2c0 .21.15.46.55.38A8 8 0 0 0 8 0z" />
      </svg>
    )
  }
  if (host === 'azure') {
    return (
      <svg viewBox="0 0 16 16" aria-hidden="true" className={cn('fill-current', className)}>
        <path d="M15 3.6v8.5l-3.5 2.9-5.4-2V15L3.1 11l8.8.7V4.2zM11.9 4.1 6.9 1v2L2.3 4.4 1 6.1v3.8l1.9.9V5.8z" />
      </svg>
    )
  }
  return <FolderGit2 aria-hidden="true" className={className} />
}
