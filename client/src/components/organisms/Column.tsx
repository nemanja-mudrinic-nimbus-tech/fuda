import { ChevronDown, ChevronsRightLeft, CircleHelp } from 'lucide-react'
import { useState } from 'react'
import { Hint } from '@/components/atoms/Hint'
import { StatusIcon } from '@/components/atoms/StatusIcon'
import { TaskCard } from '@/components/molecules/TaskCard'
import { Button } from '@/components/ui/button'
import type { Card, Column as ColumnData } from '@/lib/api'
import { cn } from '@/lib/utils'

const pageSize = 25

function columnStatus(column: ColumnData): string {
  return column.prOpen ? 'in review' : (column.statuses[0] ?? '')
}

export function Column({
  column,
  cards,
  saving,
  dropAllowed,
  isDraggable,
  onDragStart,
  onDragEnd,
  onDrop,
  collapsed,
  onToggleCollapse,
}: {
  column: ColumnData
  cards: Card[]
  saving: Set<string>
  dropAllowed: boolean
  isDraggable: (card: Card) => boolean
  onDragStart: (card: Card) => void
  onDragEnd: () => void
  onDrop: () => void
  collapsed: boolean
  onToggleCollapse: () => void
}) {
  const [limit, setLimit] = useState(pageSize)
  const [over, setOver] = useState(false)
  const status = columnStatus(column)

  if (collapsed) {
    return (
      <Hint label={`Expand ${column.name}`}>
        <button
          type="button"
          onClick={onToggleCollapse}
          className="flex w-10 shrink-0 snap-start flex-col items-center gap-3 rounded-xl bg-muted/50 py-3 text-[13px] font-medium transition-colors hover:bg-muted dark:bg-muted/30"
        >
          <StatusIcon status={status} />
          <span className="text-muted-foreground tabular-nums">{cards.length}</span>
          <span className="[writing-mode:vertical-rl]">{column.name}</span>
        </button>
      </Hint>
    )
  }

  const visible = cards.slice(0, limit)
  const hidden = cards.length - visible.length

  return (
    <section
      onDragOver={(e) => {
        if (!dropAllowed) return
        e.preventDefault()
        setOver(true)
      }}
      onDragLeave={(e) => {
        if (!e.currentTarget.contains(e.relatedTarget as Node | null)) setOver(false)
      }}
      onDrop={(e) => {
        if (!dropAllowed) return
        e.preventDefault()
        setOver(false)
        onDrop()
      }}
      className={cn(
        'group/column flex w-[85vw] shrink-0 snap-start flex-col rounded-xl bg-muted/50 sm:w-76 dark:bg-muted/30',
        over && 'ring-2 ring-primary/40',
      )}
    >
      <header className="flex h-10 items-center gap-2 px-3 text-[13px] font-medium">
        <StatusIcon status={status} />
        <span>{column.name}</span>
        <span className="text-muted-foreground tabular-nums">{cards.length}</span>
        {column.unknown && (
          <Hint label="No stage in stages.md lists this status, so it gets its own column">
            <CircleHelp className="size-3.5 text-amber-600" />
          </Hint>
        )}
        <Hint label="Collapse column">
          <Button
            variant="ghost"
            size="icon"
            aria-label={`Collapse ${column.name}`}
            onClick={onToggleCollapse}
            className="ml-auto size-6 text-muted-foreground opacity-0 transition-opacity group-hover/column:opacity-100 focus-visible:opacity-100"
          >
            <ChevronsRightLeft className="size-3.5" />
          </Button>
        </Hint>
      </header>
      <div className="flex min-h-0 flex-1 flex-col gap-1.5 overflow-y-auto px-1.5 pb-1.5">
        {cards.length === 0 ? (
          <p className="px-3 py-8 text-center text-xs text-muted-foreground/80">No tasks</p>
        ) : (
          visible.map((c) => (
            <TaskCard
              key={c.id}
              card={c}
              column={column}
              saving={saving.has(c.id)}
              draggable={isDraggable(c)}
              onDragStart={() => onDragStart(c)}
              onDragEnd={onDragEnd}
            />
          ))
        )}
        {hidden > 0 && (
          <Button
            variant="ghost"
            size="sm"
            className="h-8 shrink-0 gap-1.5 text-xs text-muted-foreground"
            onClick={() => setLimit((l) => l + pageSize * 4)}
          >
            <ChevronDown className="size-3.5" />
            Show {Math.min(hidden, pageSize * 4)} more of {hidden}
          </Button>
        )}
      </div>
    </section>
  )
}
