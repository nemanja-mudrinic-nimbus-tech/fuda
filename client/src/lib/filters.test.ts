import { describe, expect, it } from 'vitest'
import type { Card } from './api'
import {
  activeFilterCount,
  boardSearchSchema,
  countFacets,
  filterCards,
  sortCards,
  toggled,
} from './filters'

function card(id: string, overrides: Partial<Card> = {}): Card {
  return {
    id,
    title: `${id} title`,
    status: 'backlog',
    column: 'backlog',
    owners: [],
    testers: [],
    labels: [],
    prefix: id.replace(/[^A-Za-z].*$/, ''),
    added: '',
    claimed: '',
    prs: [],
    openPrs: [],
    blockedByIds: [],
    blocks: [],
    references: [],
    referencedBy: [],
    inProd: false,
    path: `docs/board/tasks/${id}.md`,
    ...overrides,
  }
}

const cards = [
  card('SS-2', {
    status: 'merged',
    owners: ['Ana'],
    labels: ['type:bug', 'area:api'],
    added: '2026-09-20',
  }),
  card('SS-10', { labels: ['type:feat', 'area:api'], added: '2026-10-01' }),
  card('SS-3', {
    owners: ['Ben'],
    blockedBy: 'SS-2',
    labels: ['type:bug', 'area:db'],
    added: '2026-09-25',
  }),
]

const ids = (list: Card[]) => list.map((c) => c.id)

describe('filterCards', () => {
  it('ORs values within a filter and ANDs label groups', () => {
    expect(ids(filterCards(cards, { label: 'type:bug,type:feat' }))).toEqual([
      'SS-2',
      'SS-10',
      'SS-3',
    ])
    expect(ids(filterCards(cards, { label: 'type:bug,area:api' }))).toEqual(['SS-2'])
  })

  it('filters by status, owner, blocked and available', () => {
    expect(ids(filterCards(cards, { status: 'merged' }))).toEqual(['SS-2'])
    expect(ids(filterCards(cards, { owner: 'Ben,Ana' }))).toEqual(['SS-2', 'SS-3'])
    expect(ids(filterCards(cards, { blocked: true }))).toEqual(['SS-3'])
    expect(ids(filterCards(cards, { available: true }))).toEqual(['SS-10'])
  })

  it('filters by added range, skipping tasks without a date', () => {
    expect(ids(filterCards(cards, { from: '2026-09-21', to: '2026-09-30' }))).toEqual(['SS-3'])
    expect(ids(filterCards([...cards, card('SS-9')], { from: '2026-01-01' }))).toHaveLength(3)
  })

  it('searches quick fields, and bodies only when text search is on', () => {
    expect(ids(filterCards(cards, { q: 'ben' }))).toEqual(['SS-3'])
    expect(ids(filterCards(cards, { q: 'rotate' }, ['SS-10']))).toEqual([])
    expect(ids(filterCards(cards, { q: 'rotate', text: true }, ['SS-10']))).toEqual(['SS-10'])
  })
})

describe('url state', () => {
  it('drops malformed params instead of failing', () => {
    expect(boardSearchSchema.parse({ status: 42, sort: 'nope', blocked: 'yes', q: '  ' })).toEqual(
      {},
    )
  })

  it('toggles list values in and out of the comma list', () => {
    expect(toggled(undefined, 'merged')).toBe('merged')
    expect(toggled('merged', 'testing')).toBe('merged,testing')
    expect(toggled('merged,testing', 'merged')).toBe('testing')
    expect(toggled('merged', 'merged')).toBeUndefined()
  })

  it('counts active filters', () => {
    expect(activeFilterCount({ status: 'a,b', blocked: true, sort: 'added', task: 'SS-1' })).toBe(3)
  })
})

describe('sortCards', () => {
  it('sorts ids naturally, or by newest added', () => {
    expect(ids(sortCards(cards, 'id'))).toEqual(['SS-2', 'SS-3', 'SS-10'])
    expect(ids(sortCards(cards, 'added'))).toEqual(['SS-10', 'SS-3', 'SS-2'])
  })
})

describe('countFacets', () => {
  it('counts each value across cards', () => {
    const counts = countFacets(cards)
    expect(counts.label.get('type:bug')).toBe(2)
    expect(counts.owner.get('Ana')).toBe(1)
    expect(counts.status.get('backlog')).toBe(2)
  })
})
