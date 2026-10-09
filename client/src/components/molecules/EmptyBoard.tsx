import { Inbox } from 'lucide-react'

export function EmptyBoard() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3 p-8 text-center">
      <div className="flex size-10 items-center justify-center rounded-full bg-muted text-muted-foreground">
        <Inbox className="size-5" />
      </div>
      <p className="text-sm font-medium">
        No Tasks yet: add files in <code className="font-mono text-xs">tasks/</code>
      </p>
    </div>
  )
}
