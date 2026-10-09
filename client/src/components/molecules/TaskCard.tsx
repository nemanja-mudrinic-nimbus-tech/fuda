import { Ban, GitPullRequest, Rocket } from 'lucide-react'
import type { ReactNode } from 'react'
import { Hint } from '@/components/atoms/Hint'
import { LabelChip } from '@/components/atoms/LabelChip'
import { labelText } from '@/lib/labels'
import { Avatar } from '@/components/atoms/PersonChip'
import { StatusIcon } from '@/components/atoms/StatusIcon'
import { useBoardSearch } from '@/hooks/useBoardSearch'
import type { Card, Column } from '@/lib/api'
import { listOf } from '@/lib/filters'
import { cn } from '@/lib/utils'

const visibleLabels = 2
const labelPriority = ['type', 'epic']
const chipClass =
  'inline-flex h-5 max-w-full items-center gap-1 rounded-full border border-border px-2 text-[11px] text-muted-foreground'

function rankLabels(labels: string[]): string[] {
  const rank = (label: string) => {
    const i = labelPriority.indexOf(labelText(label).group)
    return i === -1 ? labelPriority.length : i
  }
  return [...labels].sort((a, b) => rank(a) - rank(b))
}

function Signal({
  hint,
  icon,
  children,
}: {
  hint: ReactNode
  icon: ReactNode
  children?: ReactNode
}) {
  return (
    <Hint label={hint}>
      <span className={chipClass}>
        {icon}
        {children && <span className="truncate">{children}</span>}
      </span>
    </Hint>
  )
}

function CardSignals({ card }: { card: Card }) {
  return (
    <>
      {card.blockedBy && (
        <Signal
          hint={`Blocked by ${card.blockedBy}`}
          icon={<Ban className="size-3 shrink-0 text-amber-600" />}
        >
          <span className="block max-w-36 truncate">{card.blockedBy}</span>
        </Signal>
      )}
      {card.openPrs.map((pr) => (
        <Hint key={pr.url} label={`Open pull request #${pr.number} names this task`}>
          <a
            href={pr.url}
            target="_blank"
            rel="noreferrer"
            onClick={(e) => e.stopPropagation()}
            className={cn(chipClass, 'hover:border-violet-400 hover:text-foreground')}
          >
            <GitPullRequest className="size-3 text-violet-500" />#{pr.number}
          </a>
        </Hint>
      ))}
      {card.inProd && (
        <Signal
          hint="Its file on main is merged, testing or validated"
          icon={<Rocket className="size-3 text-emerald-600" />}
        >
          prod
        </Signal>
      )}
    </>
  )
}

export function TaskCard({
  card,
  column,
  saving,
  draggable,
  onDragStart,
  onDragEnd,
}: {
  card: Card
  column: Column
  saving: boolean
  draggable: boolean
  onDragStart: () => void
  onDragEnd: () => void
}) {
  const { search, toggle, openTask } = useBoardSearch()
  const selectedLabels = listOf(search.label)
  const ranked = rankLabels(card.labels)
  const hiddenLabels = ranked.slice(visibleLabels)
  const showStatus = column.statuses.length > 1 || column.prOpen

  return (
    <div
      role="button"
      tabIndex={0}
      draggable={draggable}
      onDragStart={(e) => {
        e.dataTransfer.effectAllowed = 'move'
        e.dataTransfer.setData('text/plain', card.id)
        onDragStart()
      }}
      onDragEnd={onDragEnd}
      aria-busy={saving}
      onClick={() => openTask(card.id)}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault()
          openTask(card.id)
        }
      }}
      className={cn(
        'group cursor-pointer rounded-lg border border-border bg-card px-3 py-2.5 text-left shadow-[0_1px_1px_rgb(0_0_0/0.03)] transition-[border-color,box-shadow]',
        'hover:border-foreground/15 hover:shadow-sm focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none',
        search.task === card.id && 'border-primary/60 ring-2 ring-primary/20',
        draggable && 'active:cursor-grabbing',
      )}
    >
      <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
        {showStatus && (
          <Hint label={`Status: ${card.status}`}>
            <StatusIcon status={card.status} className="size-3" />
          </Hint>
        )}
        <span className="font-mono tracking-tight">{card.id}</span>
        {saving && (
          <Hint label="Saving to git">
            <span
              role="status"
              aria-label="Saving"
              className="size-1.5 animate-pulse rounded-full bg-amber-500"
            />
          </Hint>
        )}
        <span className="ml-auto flex -space-x-1">
          {card.owners.map((owner) => (
            <Avatar
              key={owner}
              name={owner}
              hint={`Owner: ${owner}`}
              className="size-4.5 text-[8px] ring-2 ring-card"
            />
          ))}
        </span>
      </div>
      <p className="mt-1 text-[13px] leading-snug text-card-foreground">{card.title}</p>
      <div className="mt-2 flex flex-wrap items-center gap-1 empty:hidden">
        <CardSignals card={card} />
        {ranked.slice(0, visibleLabels).map((label) => (
          <LabelChip
            key={label}
            label={label}
            showGroup={false}
            active={selectedLabels.includes(label)}
            onClick={() => toggle('label', label)}
          />
        ))}
        {hiddenLabels.length > 0 && (
          <Hint label={hiddenLabels.join(', ')}>
            <span className="text-[11px] text-muted-foreground">+{hiddenLabels.length}</span>
          </Hint>
        )}
      </div>
    </div>
  )
}
