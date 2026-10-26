import { X } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { missingInstallUrl } from '@/lib/hosts'
import { useBoardListing } from '@/lib/queries'

export function InstallBanner() {
  const { data } = useBoardListing()
  const [closed, setClosed] = useState(false)
  const url = missingInstallUrl(data)
  if (!url || closed) return null
  return (
    <div
      role="status"
      className="flex shrink-0 items-start gap-2 border-b border-border/70 bg-muted px-4 py-2 text-sm"
    >
      <p className="min-w-0 flex-1">
        The fuda GitHub App is not installed on your account.{' '}
        <a href={url} className="font-medium underline underline-offset-2">
          Install the fuda GitHub App
        </a>
        . If you cannot, send this link to an owner of your organization:{' '}
        <code className="break-all">{url}</code>
      </p>
      <Button
        variant="ghost"
        size="icon-sm"
        aria-label="Close"
        className="shrink-0"
        onClick={() => setClosed(true)}
      >
        <X />
      </Button>
    </div>
  )
}
