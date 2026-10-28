import { Avatar } from '@/components/atoms/PersonChip'
import { LabelChip } from '@/components/atoms/LabelChip'
import { StatusIcon } from '@/components/atoms/StatusIcon'
import { useBoardSearch } from '@/hooks/useBoardSearch'
import type { Card, Column } from '@/lib/api'
import { listOf } from '@/lib/filters'
import { columnStatus } from '@/lib/moves'
import { labelText } from '@/lib/labels'
import { cn } from '@/lib/utils'

function Row({ card }: { card: Card }) {
  const { search, toggle, openTask } = useBoardSearch()
  const selected = listOf(search.label)
  const labels = card.labels
    .filter((l) => ['type', 'epic'].includes(labelText(l).group))
    .slice(0, 2)
  return (
    <li>
      <div
        role="button"
        tabIndex={0}
        onClick={() => openTask(card.id)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            openTask(card.id)
          }
        }}
        className={cn(
          'flex h-10 cursor-pointer items-center gap-3 border-b border-border/60 px-3 text-[13px] transition-colors hover:bg-accent/60 focus-visible:bg-accent focus-visible:outline-none',
          search.task === card.id && 'bg-accent',
        )}
      >
        <StatusIcon status={card.status} className="size-3" />
        <span className="w-16 shrink-0 font-mono text-xs text-muted-foreground">{card.id}</span>
        <span className="min-w-0 flex-1 truncate">{card.title}</span>
        <span className="hidden shrink-0 items-center gap-1 md:flex">
          {labels.map((l) => (
            <LabelChip
              key={l}
              label={l}
              showGroup={false}
              active={selected.includes(l)}
              onClick={() => toggle('label', l)}
            />
          ))}
        </span>
        <span className="hidden w-20 shrink-0 text-right text-xs text-muted-foreground tabular-nums sm:block">
          {card.added}
        </span>
        <span className="flex w-12 shrink-0 justify-end -space-x-1">
          {card.owners.map((o) => (
            <Avatar
              key={o}
              name={o}
              hint={`Owner: ${o}`}
              className="size-5 ring-2 ring-background"
            />
          ))}
        </span>
      </div>
    </li>
  )
}

export function ListView({ columns, cards }: { columns: Column[]; cards: Card[] }) {
  return (
    <div className="min-h-0 flex-1 overflow-y-auto px-3 pb-6 sm:px-4">
      <div className="mx-auto max-w-6xl overflow-hidden rounded-xl border border-border bg-card shadow-xs">
        {columns.map((column) => {
          const rows = cards.filter((c) => c.column === column.id)
          if (rows.length === 0) return null
          return (
            <section key={column.id}>
              <h2 className="sticky top-0 z-10 flex h-9 items-center gap-2 border-b border-border/60 bg-muted/80 px-3 text-xs font-medium backdrop-blur">
                <StatusIcon status={columnStatus(column)} className="size-3" />
                {column.name}
                <span className="text-muted-foreground tabular-nums">{rows.length}</span>
              </h2>
              <ul>
                {rows.map((card) => (
                  <Row key={card.id} card={card} />
                ))}
              </ul>
            </section>
          )
        })}
      </div>
    </div>
  )
}
