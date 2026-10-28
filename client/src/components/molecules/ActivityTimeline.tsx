import { Avatar } from '@/components/atoms/PersonChip'
import type { Activity } from '@/lib/api'
import { resultTone, type ResultTone } from '@/lib/activity'
import { cn } from '@/lib/utils'

const toneClass: Record<ResultTone, string> = {
  pass: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400',
  warn: 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-400',
  fail: 'border-red-500/30 bg-red-500/10 text-red-700 dark:text-red-400',
  other: 'border-border bg-muted text-muted-foreground',
}

export function ResultBadge({ result }: { result: string }) {
  const tone = resultTone(result)
  return (
    <span
      data-tone={tone}
      className={cn(
        'inline-flex h-5 items-center rounded-full border px-2 font-mono text-[11px] font-medium',
        toneClass[tone],
      )}
    >
      {result}
    </span>
  )
}

function Entry({ entry }: { entry: Activity }) {
  if (!entry.actor) {
    return (
      <p className="text-[13px] text-muted-foreground [overflow-wrap:anywhere]">{entry.text}</p>
    )
  }
  return (
    <div className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-[13px]">
      <time className="font-mono text-[11px] text-muted-foreground">{entry.when}</time>
      <span className="inline-flex items-center gap-1.5">
        <Avatar name={entry.actor} />
        <span className="font-medium">{entry.actor}</span>
      </span>
      <span>{entry.event}</span>
      {entry.result && <ResultBadge result={entry.result} />}
      {entry.usage && <span className="text-xs text-muted-foreground">{entry.usage}</span>}
    </div>
  )
}

export function ActivityTimeline({ entries }: { entries: Activity[] }) {
  if (entries.length === 0) return null
  const newestFirst = entries.map((entry, index) => ({ entry, index })).reverse()
  return (
    <section aria-label="Activity">
      <h2 className="mb-2 text-sm font-semibold">Activity</h2>
      <ol className="space-y-2 border-l border-border pl-4">
        {newestFirst.map(({ entry, index }) => (
          <li
            key={index}
            className="relative before:absolute before:top-2 before:-left-[1.3rem] before:size-2 before:rounded-full before:bg-border"
          >
            <Entry entry={entry} />
          </li>
        ))}
      </ol>
    </section>
  )
}
