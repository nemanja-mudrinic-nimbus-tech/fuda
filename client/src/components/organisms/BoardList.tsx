import { useEffect } from 'react'
import { HostIcon } from '@/components/atoms/HostIcon'
import { Button } from '@/components/ui/button'
import { loginPath } from '@/lib/boardPath'
import { hostLabel, logOut, onlyBoard } from '@/lib/hosts'
import { useBoardListing } from '@/lib/queries'

export function BoardList() {
  const { data, error } = useBoardListing()
  const only = data && onlyBoard(data)

  useEffect(() => {
    if (only) window.location.replace(only.path)
  }, [only])

  if (error) {
    return (
      <p className="p-6 text-sm text-muted-foreground">Could not list Boards: {error.message}</p>
    )
  }
  if (!data || only) {
    return <p className="p-6 text-sm text-muted-foreground">Loading Boards…</p>
  }
  const { boards, hosts } = data
  if (boards.length === 0 && hosts.every((host) => host.loggedIn)) {
    return (
      <p className="p-6 text-sm text-muted-foreground">
        No Boards yet. Create a repository whose name starts with <code>fuda-</code> (on GitHub,
        install the fuda GitHub App on it).
      </p>
    )
  }
  return (
    <main className="mx-auto w-full max-w-md p-6">
      <h1 className="mb-3 text-base font-semibold">Boards</h1>
      <ul className="flex flex-col gap-1">
        {boards.map((board) => (
          <li key={board.path}>
            <a
              href={board.path}
              className="flex items-center gap-2 rounded-md border border-border bg-card px-3 py-2 text-sm hover:border-foreground/20"
            >
              <HostIcon host={board.host} className="size-4 shrink-0" />
              <span className="truncate">{board.repo}</span>
            </a>
          </li>
        ))}
      </ul>
      <div className="mt-4 flex flex-wrap gap-2">
        {hosts.map((host) =>
          host.loggedIn ? (
            <Button
              key={host.host}
              variant="outline"
              size="sm"
              onClick={() => void logOut(host.host)}
            >
              Log out of {hostLabel(host.host)}
            </Button>
          ) : (
            <Button key={host.host} size="sm" asChild>
              <a href={loginPath('/', host.login)}>Log in to {hostLabel(host.host)}</a>
            </Button>
          ),
        )}
        {hosts.length > 1 && hosts.some((host) => host.loggedIn) && (
          <Button variant="ghost" size="sm" onClick={() => void logOut()}>
            Log out of all
          </Button>
        )}
      </div>
    </main>
  )
}
