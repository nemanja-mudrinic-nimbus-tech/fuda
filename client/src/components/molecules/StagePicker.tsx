import { ArrowRightLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Column } from '@/lib/api'

export function StagePicker({
  targets,
  saving,
  onMove,
}: {
  targets: Column[]
  saving: boolean
  onMove: (column: Column) => void
}) {
  if (targets.length === 0) return null
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="xs" disabled={saving} className="text-muted-foreground">
          <ArrowRightLeft className="size-3.5" />
          {saving ? 'Saving…' : 'Move to'}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start">
        {targets.map((column) => (
          <DropdownMenuItem key={column.id} onSelect={() => onMove(column)}>
            {column.name}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
