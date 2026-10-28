import { format, parseISO } from 'date-fns'
import { FilterPill, PillShell } from '@/components/molecules/FilterPill'
import { Button } from '@/components/ui/button'
import { useBoardSearch } from '@/hooks/useBoardSearch'
import type { FilterGroup } from '@/hooks/useFilterGroups'
import { listOf, NO_OWNER } from '@/lib/filters'

function day(value?: string): string {
  return value ? format(parseISO(value), 'd MMM') : '…'
}

function FlagPill({ label, onRemove }: { label: string; onRemove: () => void }) {
  return (
    <PillShell onRemove={onRemove} removeLabel={`Remove ${label}`}>
      <span className="flex items-center px-2 font-medium">{label}</span>
    </PillShell>
  )
}

function ownerNames(group: FilterGroup, clearNames: () => void): FilterGroup {
  return {
    ...group,
    selected: group.selected.filter((v) => v !== NO_OWNER),
    clear: clearNames,
  }
}

export function ActiveFilters({ groups }: { groups: FilterGroup[] }) {
  const { search, update, toggle, clear } = useBoardSearch()
  const noOwner = listOf(search.owner).includes(NO_OWNER)
  const active = groups
    .map((g) =>
      g.id === 'owner' && noOwner ? ownerNames(g, () => update({ owner: NO_OWNER })) : g,
    )
    .filter((g) => g.selected.length > 0)
  const dated = Boolean(search.from || search.to)
  if (
    active.length === 0 &&
    !noOwner &&
    !dated &&
    !search.blocked &&
    !search.available &&
    !search.q
  )
    return null

  return (
    <>
      {active.map((g) => (
        <FilterPill key={g.id} group={g} />
      ))}
      {noOwner && <FlagPill label="No owner" onRemove={() => toggle('owner', NO_OWNER)} />}
      {dated && (
        <FlagPill
          label={`Added ${day(search.from)} – ${day(search.to)}`}
          onRemove={() => update({ from: undefined, to: undefined })}
        />
      )}
      {search.blocked && (
        <FlagPill label="Blocked" onRemove={() => update({ blocked: undefined })} />
      )}
      {search.available && (
        <FlagPill label="Available" onRemove={() => update({ available: undefined })} />
      )}
      <Button
        variant="ghost"
        size="sm"
        className="h-7 px-2 text-xs text-muted-foreground"
        onClick={clear}
      >
        Clear
      </Button>
    </>
  )
}
