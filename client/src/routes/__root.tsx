import type { QueryClient } from '@tanstack/react-query'
import { createRootRouteWithContext, Outlet } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { InstallBanner } from '@/components/molecules/InstallBanner'
import { AppHeader } from '@/components/organisms/AppHeader'
import { Toaster } from '@/components/ui/sonner'
import { TooltipProvider } from '@/components/ui/tooltip'
import { useTheme } from '@/hooks/useTheme'
import { ApiError } from '@/lib/api'
import { installUrl } from '@/lib/hosts'
import { useBoard, useBoardListing } from '@/lib/queries'

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  component: RootLayout,
})

function RootLayout() {
  const { data, error } = useBoard()
  const install = installUrl(useBoardListing().data)
  const { resolved } = useTheme()
  const client = useQueryClient()
  const head = data?.sync.develop.sha

  useEffect(() => {
    if (!head) return
    for (const key of ['task', 'search', 'archive', 'doc']) {
      void client.invalidateQueries({ queryKey: [key] })
    }
  }, [head, client])

  useEffect(() => {
    document.title = data?.title ? `${data.title} · fuda` : 'fuda'
  }, [data?.title])

  return (
    <TooltipProvider delayDuration={300}>
      <div className="flex h-svh flex-col text-foreground">
        <AppHeader board={data} />
        <InstallBanner />
        {error instanceof ApiError && error.status === 404 ? (
          <p className="p-6 text-sm text-muted-foreground">
            No Board at this address, or you cannot see it. The repository name must start with{' '}
            <code>fuda-</code>, on GitHub the fuda GitHub App must be installed on it, and your
            account must have access to it.
            {install && (
              <>
                {' '}
                <a href={install} className="underline underline-offset-2">
                  Install the fuda GitHub App
                </a>
                .
              </>
            )}
          </p>
        ) : error instanceof ApiError && error.status === 403 ? (
          <p className="p-6 text-sm text-muted-foreground">
            The host says you have no access to this Board. Ask the repository owner to give you
            access.
          </p>
        ) : (
          <Outlet />
        )}
      </div>
      <Toaster theme={resolved} position="bottom-right" richColors closeButton />
    </TooltipProvider>
  )
}
