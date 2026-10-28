import { z } from 'zod'
import type { Card } from './api'

const optionalText = z.string().trim().min(1).optional().catch(undefined)
const flag = z.boolean().optional().catch(undefined)

export const boardSearchSchema = z.object({
  status: optionalText,
  owner: optionalText,
  tester: optionalText,
  label: optionalText,
  prefix: optionalText,
  blocked: flag,
  available: flag,
  q: optionalText,
  text: flag,
  from: optionalText,
  to: optionalText,
  sort: z.enum(['id', 'added']).optional().catch(undefined),
  view: z.enum(['board', 'list']).optional().catch(undefined),
  task: optionalText,
})

export type BoardSearch = z.infer<typeof boardSearchSchema>

export const NO_OWNER = 'none'

export type ListKey = 'status' | 'owner' | 'tester' | 'label' | 'prefix'

export function listOf(value: string | undefined): string[] {
  return value ? value.split(',').filter(Boolean) : []
}

export function toggled(value: string | undefined, item: string): string | undefined {
  const list = listOf(value)
  const next = list.includes(item) ? list.filter((v) => v !== item) : [...list, item]
  return next.length ? next.join(',') : undefined
}

export function isAvailable(card: Card): boolean {
  return (
    card.status.trim().toLowerCase() === 'backlog' && card.owners.length === 0 && !card.blockedBy
  )
}

function anyOf(selected: string[], values: string[]): boolean {
  return selected.length === 0 || selected.some((s) => values.includes(s))
}

function ownersMatch(selected: string[], owners: string[]): boolean {
  return anyOf(selected, owners) || (owners.length === 0 && selected.includes(NO_OWNER))
}

function labelsMatch(selected: string[], labels: string[]): boolean {
  const byGroup = new Map<string, string[]>()
  for (const label of selected) {
    const group = label.split(':')[0]
    byGroup.set(group, [...(byGroup.get(group) ?? []), label])
  }
  return [...byGroup.values()].every((wanted) => wanted.some((l) => labels.includes(l)))
}

function quickText(card: Card): string {
  return [card.id, card.title, ...card.owners, ...card.testers, ...card.labels]
    .join(' ')
    .toLowerCase()
}

export function filterCards(cards: Card[], search: BoardSearch, textMatches?: string[]): Card[] {
  const statuses = listOf(search.status)
  const owners = listOf(search.owner)
  const testers = listOf(search.tester)
  const labels = listOf(search.label)
  const prefixes = listOf(search.prefix)
  const q = search.q?.toLowerCase()
  const bodyHits = new Set(search.text ? (textMatches ?? []) : [])

  return cards.filter(
    (card) =>
      anyOf(statuses, [card.status]) &&
      ownersMatch(owners, card.owners) &&
      anyOf(testers, card.testers) &&
      anyOf(prefixes, [card.prefix]) &&
      labelsMatch(labels, card.labels) &&
      (!search.blocked || !!card.blockedBy) &&
      (!search.available || isAvailable(card)) &&
      (!search.from || (card.added !== '' && card.added >= search.from)) &&
      (!search.to || (card.added !== '' && card.added <= search.to)) &&
      (!q || quickText(card).includes(q) || bodyHits.has(card.id)),
  )
}

const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' })

export function sortCards(cards: Card[], sort: BoardSearch['sort']): Card[] {
  const sorted = [...cards]
  if (sort === 'added') {
    return sorted.sort((a, b) => b.added.localeCompare(a.added) || collator.compare(a.id, b.id))
  }
  return sorted.sort((a, b) => collator.compare(a.id, b.id))
}

export function activeFilterCount(search: BoardSearch): number {
  const lists = (['status', 'owner', 'tester', 'label', 'prefix'] as const).reduce(
    (n, key) => n + listOf(search[key]).length,
    0,
  )
  return (
    lists +
    [search.blocked, search.available, search.q, search.from, search.to].filter(Boolean).length
  )
}

export type FacetCounts = {
  status: Map<string, number>
  owner: Map<string, number>
  tester: Map<string, number>
  label: Map<string, number>
  prefix: Map<string, number>
}

export function countFacets(cards: Card[]): FacetCounts {
  const counts: FacetCounts = {
    status: new Map(),
    owner: new Map(),
    tester: new Map(),
    label: new Map(),
    prefix: new Map(),
  }
  const add = (map: Map<string, number>, key: string) => map.set(key, (map.get(key) ?? 0) + 1)
  for (const card of cards) {
    add(counts.status, card.status)
    add(counts.prefix, card.prefix)
    card.owners.forEach((o) => add(counts.owner, o))
    if (card.owners.length === 0) add(counts.owner, NO_OWNER)
    card.testers.forEach((t) => add(counts.tester, t))
    card.labels.forEach((l) => add(counts.label, l))
  }
  return counts
}
