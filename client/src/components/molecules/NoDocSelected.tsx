import { FileText } from 'lucide-react'

export function NoDocSelected() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3 p-8 text-center">
      <div className="flex size-10 items-center justify-center rounded-full bg-muted text-muted-foreground">
        <FileText className="size-5" />
      </div>
      <div>
        <p className="text-sm font-medium">No doc selected</p>
        <p className="mt-1 text-xs text-muted-foreground">Pick a doc from the tree to read it.</p>
      </div>
    </div>
  )
}
