export type ResultTone = 'pass' | 'warn' | 'fail' | 'other'

const tones: Record<string, ResultTone> = {
  PASS: 'pass',
  READY: 'pass',
  COMPLETE: 'pass',
  auto_merge: 'pass',
  CHANGES: 'warn',
  GAPS: 'warn',
  needs_human: 'warn',
  BLOCKED: 'fail',
  FAILED: 'fail',
  HIT: 'fail',
}

const prefixed = /^\S+: (.+)$/

export function resultWord(result: string): string {
  return prefixed.exec(result)?.[1] ?? result
}

export function resultTone(result: string): ResultTone {
  const firstWord = resultWord(result).split(/\s+/)[0]
  return Object.hasOwn(tones, firstWord) ? tones[firstWord] : 'other'
}

const activityTime = /^(\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2})$/
const minute = 60_000
const hour = 60 * minute
const day = 24 * hour

export function relativeTime(when: string, now: number): string {
  const parts = activityTime.exec(when)
  if (!parts) return when
  const [year, month, date, hours, minutes] = parts.slice(1).map(Number)
  const at = new Date(year, month - 1, date, hours, minutes)
  const ago = now - at.getTime()
  if (ago < minute) return 'just now'
  if (ago < hour) return `${Math.floor(ago / minute)} min ago`
  if (ago < day) return `${Math.floor(ago / hour)} h ago`
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  if (at.toDateString() === yesterday.toDateString()) return 'yesterday'
  return when.slice(0, 10)
}

export type DetailKind = 'path' | 'url' | 'text'

export function detailKind(detail: string): DetailKind {
  if (/^https?:\/\/\S+$/.test(detail)) return 'url'
  if (detail.startsWith('~/') || detail.startsWith('/')) return 'path'
  return 'text'
}
