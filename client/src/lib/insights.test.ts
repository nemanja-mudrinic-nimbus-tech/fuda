import { describe, expect, it } from 'vitest'
import type { Card } from './api'
import {
  activity,
  countBy,
  insightsSearchSchema,
  isOpen,
  labelsOf,
  periodOf,
  topSlices,
} from './insights'

function card(id: string, overrides: Partial<Card> = {}): Card {
  return {
    id,
    title: id,
    status: 'backlog',
    column: 'backlog',
    owners: [],
    testers: [],
    labels: [],
    prefix: 'X',
    added: '',
    claimed: '',
    prs: [],
    openPrs: [],
    blockedByIds: [],
    blocks: [],
    references: [],
    referencedBy: [],
    inProd: false,
    path: '',
    ...overrides,
  }
}

describe('activity', () => {
  const today = new Date('2026-10-03T12:00:00')
  const cards = [
    card('A', { added: '2026-09-29', claimed: '2026-10-02' }),
    card('B', { added: '2026-09-22' }),
    card('C', { added: '2026-08-01' }),
  ]

  it('buckets a week range into ISO weeks ending this week', () => {
    const { unit, buckets } = activity(cards, periodOf({ range: '4w' }, today))
    expect(unit).toBe('week')
    expect(buckets.map((b) => b.start)).toEqual([
      '2026-09-07',
      '2026-09-14',
      '2026-09-21',
      '2026-09-28',
    ])
    expect(buckets.map((b) => [b.created, b.claimed])).toEqual([
      [0, 0],
      [0, 0],
      [1, 0],
      [1, 1],
    ])
  })

  it('buckets the last 7 days per day', () => {
    const { unit, buckets } = activity(cards, periodOf({ range: '7d' }, today))
    expect(unit).toBe('day')
    expect(buckets.map((b) => b.start)).toEqual([
      '2026-09-27',
      '2026-09-28',
      '2026-09-29',
      '2026-09-30',
      '2026-10-01',
      '2026-10-02',
      '2026-10-03',
    ])
    expect(buckets.reduce((n, b) => n + b.created, 0)).toBe(1)
  })

  it('uses a custom range inclusive of both ends', () => {
    const period = periodOf({ from: '2026-08-01', to: '2026-09-22' }, today)
    const { unit, buckets } = activity(cards, period)
    expect(unit).toBe('week')
    expect(buckets.reduce((n, b) => n + b.created, 0)).toBe(2)
    expect(period.label).toBe('1 Aug 2026 – 22 Sep 2026')
  })

  it('drops malformed dates from the URL', () => {
    expect(insightsSearchSchema.parse({ from: '2026-13-45', to: 'soon' })).toEqual({})
  })
})

describe('breakdowns', () => {
  const cards = [
    card('A', { labels: ['epic:billing', 'type:bug'] }),
    card('B', { labels: ['epic:billing'] }),
    card('C', { labels: ['type:feat'], status: 'merged' }),
  ]
  it('counts by label group with None for unlabeled', () => {
    expect(countBy(cards, labelsOf('epic'))).toEqual(
      new Map([
        ['billing', 2],
        ['None', 1],
      ]),
    )
  })
  it('keeps the top slices and folds the rest into Other', () => {
    const counts = new Map([
      ['a', 5],
      ['b', 4],
      ['c', 3],
      ['d', 1],
    ])
    expect(topSlices(counts, 3)).toEqual([
      { key: 'a', count: 5 },
      { key: 'b', count: 4 },
      { key: 'Other', count: 4 },
    ])
  })
  it('treats merged and later as not open', () => {
    expect(cards.filter(isOpen).map((c) => c.id)).toEqual(['A', 'B'])
  })
})
