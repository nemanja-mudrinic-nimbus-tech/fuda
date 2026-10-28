import { Check, Copy } from 'lucide-react'
import { useState } from 'react'
import { Avatar } from '@/components/atoms/PersonChip'
import type { Activity } from '@/lib/api'
import { detailKind, relativeTime, resultTone, resultWord, type ResultTone } from '@/lib/activity'
import { cn } from '@/lib/utils'

const toneClass: Record<ResultTone, string> = {
  pass: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400',
  warn: 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-400',
  fail: 'border-red-500/30 bg-red-500/10 text-red-700 dark:text-red-400',
  other: 'border-border bg-muted text-muted-foreground',
}

const visibleDetails = 3

export function ResultBadge({ result, className }: { result: string; className?: string }) {
  const tone = resultTone(result)
  return (
    <span
      data-tone={tone}
      title={result}
      className={cn(
        'inline-flex h-5 shrink-0 items-center rounded-full border px-2 font-mono text-[11px] font-medium',
        toneClass[tone],
        className,
      )}
    >
      {resultWord(result)}
    </span>
  )
}

function CopyPath({ path }: { path: string }) {
  const [copied, setCopied] = useState(false)
  const copy = () =>
    void navigator.clipboard.writeText(path).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    })
  return (
    <span className="inline-flex min-w-0 items-center gap-1">
      <code className="font-mono text-xs [overflow-wrap:anywhere]">{path}</code>
      <button
        type="button"
        onClick={copy}
        aria-label={copied ? 'Copied' : `Copy ${path}`}
        className="shrink-0 rounded p-0.5 text-muted-foreground hover:bg-muted hover:text-foreground"
      >
        {copied ? <Check className="size-3" /> : <Copy className="size-3" />}
      </button>
    </span>
  )
}

function Detail({ detail }: { detail: string }) {
  switch (detailKind(detail)) {
    case 'url':
      return (
        <a
          href={detail}
          target="_blank"
          rel="noreferrer"
          className="text-primary underline-offset-2 [overflow-wrap:anywhere] hover:underline"
        >
          {detail}
        </a>
      )
    case 'path':
      return <CopyPath path={detail} />
    default:
      return <span className="[overflow-wrap:anywhere]">{detail}</span>
  }
}

function Details({ details }: { details: string[] }) {
  const [expanded, setExpanded] = useState(false)
  const hidden = details.length - visibleDetails
  const shown = expanded || hidden <= 0 ? details : details.slice(0, visibleDetails)
  return (
    <div className="mt-1 text-xs text-muted-foreground">
      <ul className="list-disc space-y-0.5 pl-4">
        {shown.map((detail, index) => (
          <li key={index}>
            <Detail detail={detail} />
          </li>
        ))}
      </ul>
      {hidden > 0 && (
        <button
          type="button"
          onClick={() => setExpanded(!expanded)}
          className="mt-0.5 text-xs font-medium text-foreground/70 hover:text-foreground"
        >
          {expanded ? 'show less' : `show ${hidden} more`}
        </button>
      )}
    </div>
  )
}

function Entry({ entry, now }: { entry: Activity; now: number }) {
  const details = entry.details ?? []
  if (!entry.actor) {
    return (
      <>
        <p className="text-[13px] text-muted-foreground [overflow-wrap:anywhere]">{entry.text}</p>
        {details.length > 0 && <Details details={details} />}
      </>
    )
  }
  return (
    <>
      <div className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-[13px]">
        <span className="inline-flex items-center gap-1.5">
          <Avatar name={entry.actor} />
          <span className="font-medium">{entry.actor}</span>
        </span>
        <span>{entry.event}</span>
        {entry.result && <ResultBadge result={entry.result} />}
        <time
          dateTime={entry.when}
          title={entry.when}
          className="text-[11px] text-muted-foreground"
        >
          {relativeTime(entry.when ?? '', now)}
        </time>
        {entry.usage && (
          <span className="ml-auto text-xs text-muted-foreground">{entry.usage}</span>
        )}
      </div>
      {details.length > 0 && <Details details={details} />}
    </>
  )
}

export function ActivityTimeline({ entries, now }: { entries: Activity[]; now: number }) {
  if (entries.length === 0) return null
  const newestFirst = entries.map((entry, index) => ({ entry, index })).reverse()
  return (
    <section aria-label="Activity">
      <h2 className="mb-2 text-sm font-semibold">Activity</h2>
      <ol className="space-y-2.5 border-l border-border pl-4">
        {newestFirst.map(({ entry, index }) => (
          <li
            key={index}
            className="relative before:absolute before:top-2 before:-left-[1.3rem] before:size-2 before:rounded-full before:bg-border"
          >
            <Entry entry={entry} now={now} />
          </li>
        ))}
      </ol>
    </section>
  )
}
