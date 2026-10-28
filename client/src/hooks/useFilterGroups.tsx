import { UserRound } from 'lucide-react'
import { useMemo, type ReactNode } from 'react'
import { Avatar } from '@/components/atoms/PersonChip'
import { StatusIcon } from '@/components/atoms/StatusIcon'
import { useBoardSearch } from '@/hooks/useBoardSearch'
import type { Card, Column, Facets } from '@/lib/api'
import { countFacets, listOf, NO_OWNER } from '@/lib/filters'

export type FilterValue = { value: string; label: ReactNode; text: string; count: number }

export type FilterGroup = {
  id: string
  title: string
  plural: string
  values: FilterValue[]
  selected: string[]
  toggle: (value: string) => void
  clear: () => void
}

const groupTitles: Record<string, [string, string]> = {
  type: ['Type', 'types'],
  epic: ['Epic', 'epics'],
  area: ['Area', 'areas'],
  flag: ['Flag', 'flags'],
  label: ['Label', 'labels'],
}

function titleOf(group: string): [string, string] {
  return groupTitles[group] ?? [group.charAt(0).toUpperCase() + group.slice(1), group]
}

function person(name: string): ReactNode {
  return (
    <span className="inline-flex items-center gap-2">
      <Avatar name={name} className="size-4.5 text-[8px]" />
      {name}
    </span>
  )
}

function noOwner(count: number): FilterValue {
  return {
    value: NO_OWNER,
    text: 'No owner',
    count,
    label: (
      <span className="inline-flex items-center gap-2">
        <UserRound className="size-5 shrink-0 rounded-full border border-dashed border-muted-foreground/60 p-0.5 text-muted-foreground" />
        No owner
      </span>
    ),
  }
}

export function useFilterGroups(facets: Facets, columns: Column[], cards: Card[]): FilterGroup[] {
  const { search, toggle, update } = useBoardSearch()
  const counts = useMemo(() => countFacets(cards), [cards])
  const labels = listOf(search.label)

  const labelGroups: FilterGroup[] = facets.labelGroups.map((g) => {
    const prefix = `${g.name}:`
    const [title, plural] = titleOf(g.name)
    return {
      id: `label:${g.name}`,
      title,
      plural,
      values: g.values.map((v) => ({
        value: prefix + v,
        text: v,
        count: counts.label.get(prefix + v) ?? 0,
        label: (
          <span className="inline-flex items-center gap-2">
            <span
              className="size-2 shrink-0 rounded-full"
              style={{ backgroundColor: facets.labelColors[prefix + v] }}
            />
            {v}
          </span>
        ),
      })),
      selected: labels.filter((l) => l.startsWith(prefix)),
      toggle: (v) => toggle('label', v),
      clear: () =>
        update({ label: labels.filter((l) => !l.startsWith(prefix)).join(',') || undefined }),
    }
  })

  const people = (
    key: 'owner' | 'tester',
    title: string,
    plural: string,
    names: string[],
    leading: FilterValue[] = [],
  ): FilterGroup => ({
    id: key,
    title,
    plural,
    values: [
      ...leading,
      ...names.map((n) => ({
        value: n,
        text: n,
        label: person(n),
        count: counts[key].get(n) ?? 0,
      })),
    ],
    selected: listOf(search[key]),
    toggle: (v) => toggle(key, v),
    clear: () => update({ [key]: undefined }),
  })

  const statusNeeded = columns.some((c) => c.statuses.length > 1 || c.unknown)

  const groups: FilterGroup[] = [
    ...labelGroups.filter((g) => g.id === 'label:type' || g.id === 'label:epic'),
    people('owner', 'Owner', 'owners', facets.people, [noOwner(counts.owner.get(NO_OWNER) ?? 0)]),
    ...labelGroups.filter((g) => g.id !== 'label:type' && g.id !== 'label:epic'),
    people('tester', 'Tester', 'testers', facets.testers),
    {
      id: 'status',
      title: 'Status',
      plural: 'statuses',
      values: (statusNeeded ? facets.statuses : []).map((s) => ({
        value: s,
        text: s,
        count: counts.status.get(s) ?? 0,
        label: (
          <span className="inline-flex items-center gap-2">
            <StatusIcon status={s} className="size-3" />
            {s}
          </span>
        ),
      })),
      selected: listOf(search.status),
      toggle: (v) => toggle('status', v),
      clear: () => update({ status: undefined }),
    },
    {
      id: 'prefix',
      title: 'Id prefix',
      plural: 'prefixes',
      values: (facets.prefixes.length > 1 ? facets.prefixes : []).map((p) => ({
        value: p,
        text: p,
        count: counts.prefix.get(p) ?? 0,
        label: <span className="font-mono">{p}</span>,
      })),
      selected: listOf(search.prefix),
      toggle: (v) => toggle('prefix', v),
      clear: () => update({ prefix: undefined }),
    },
  ]
  return groups.filter((g) => g.values.length > 0)
}
