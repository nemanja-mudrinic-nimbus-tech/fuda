import { CircleUser } from 'lucide-react'
import { HostIcon } from '@/components/atoms/HostIcon'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { loginPath } from '@/lib/boardPath'
import { hostLabel, logOut, showLogOutAll } from '@/lib/hosts'
import { useBoardListing } from '@/lib/queries'

export function AccountMenu() {
  const { data } = useBoardListing()
  if (!data || data.hosts.length === 0) return null
  const returnTo = window.location.pathname + window.location.search
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="size-7" aria-label="Account">
          <CircleUser className="size-3.5" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-64">
        {data.hosts.map((host) => (
          <div key={host.host} className="flex items-center gap-2 px-2 py-1.5 text-sm">
            <HostIcon host={host.host} className="size-3.5 shrink-0" />
            <div className="min-w-0 flex-1">
              <div className="truncate">{hostLabel(host.host)}</div>
              <div className="text-xs text-muted-foreground">
                {host.loggedIn ? 'Logged in' : 'Logged out'}
              </div>
            </div>
            {host.loggedIn ? (
              <Button variant="ghost" size="sm" onClick={() => void logOut(host.host)}>
                Log out
              </Button>
            ) : (
              <Button variant="ghost" size="sm" asChild>
                <a href={loginPath(returnTo, host.login)}>Log in</a>
              </Button>
            )}
          </div>
        ))}
        {showLogOutAll(data.hosts) && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem onSelect={() => void logOut()}>Log out of all</DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
