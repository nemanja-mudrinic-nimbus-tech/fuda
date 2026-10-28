import { describe, expect, it } from 'vitest'
import { detailKind, relativeTime, resultTone, resultWord } from './activity'

describe('resultTone', () => {
  it.each([
    ['PASS', 'pass'],
    ['READY', 'pass'],
    ['COMPLETE', 'pass'],
    ['auto_merge', 'pass'],
    ['CHANGES', 'warn'],
    ['GAPS', 'warn'],
    ['needs_human', 'warn'],
    ['BLOCKED', 'fail'],
    ['FAILED', 'fail'],
    ['HIT', 'fail'],
    ['pass', 'other'],
    ['SKIPPED', 'other'],
    ['toString', 'other'],
    ['VERDICT: PASS', 'pass'],
    ['VERDICT: CHANGES', 'warn'],
    ['auto_merge 0.92', 'pass'],
  ])('%s is %s', (result, tone) => {
    expect(resultTone(result)).toBe(tone)
  })
})

describe('resultWord', () => {
  it.each([
    ['VERDICT: PASS', 'PASS'],
    ['STATUS: READY', 'READY'],
    ['COMPLETE', 'COMPLETE'],
    ['auto_merge 0.92', 'auto_merge 0.92'],
    ['decision: auto_merge 0.92', 'auto_merge 0.92'],
    ['two words: PASS', 'two words: PASS'],
  ])('%s shows %s', (result, word) => {
    expect(resultWord(result)).toBe(word)
  })
})

describe('relativeTime', () => {
  const now = new Date(2026, 9, 28, 14, 30).getTime()
  it.each([
    ['2026-10-28 14:30', 'just now'],
    ['2026-10-28 14:45', 'just now'],
    ['2026-10-28 14:18', '12 min ago'],
    ['2026-10-28 11:29', '3 h ago'],
    ['2026-10-27 15:00', '23 h ago'],
    ['2026-10-27 09:12', 'yesterday'],
    ['2026-10-26 23:59', '2026-10-26'],
    ['2026-09-01 08:00', '2026-09-01'],
    ['not a time', 'not a time'],
  ])('%s is %s', (when, text) => {
    expect(relativeTime(when, now)).toBe(text)
  })
})

describe('detailKind', () => {
  it.each([
    ['~/runs/F-8/evidence.md', 'path'],
    ['/tmp/check.log', 'path'],
    ['https://example.com/pr/5', 'url'],
    ['http://localhost:8080', 'url'],
    ['see https://example.com', 'text'],
    ['next: code review', 'text'],
  ])('%s is a %s', (detail, kind) => {
    expect(detailKind(detail)).toBe(kind)
  })
})
