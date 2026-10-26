import { createFileRoute } from '@tanstack/react-router'
import { useMemo } from 'react'
import { EmptyBoard } from '@/components/molecules/EmptyBoard'
import { NoMatches } from '@/components/molecules/NoMatches'
import { Board } from '@/components/organisms/Board'
import { BoardList } from '@/components/organisms/BoardList'
import { ListView } from '@/components/organisms/ListView'
import { FilterBar } from '@/components/organisms/FilterBar'
import { TaskSheet } from '@/components/organisms/TaskSheet'
import { boardSearchSchema, filterCards, sortCards } from '@/lib/filters'
import { BOARD_PREFIX } from '@/lib/boardPath'
import { useBoard, useTextSearch } from '@/lib/queries'

export const Route = createFileRoute('/')({
  validateSearch: boardSearchSchema,
  component: BoardPage,
})

function BoardPage() {
  const search = Route.useSearch()
  const { data, error } = useBoard()
  const { data: textMatches } = useTextSearch(search.q, !!search.text)

  const cards = useMemo(
    () => (data ? sortCards(filterCards(data.cards, search, textMatches), search.sort) : []),
    [data, search, textMatches],
  )
  const taskIds = useMemo(() => data?.cards.map((c) => c.id) ?? [], [data])

  if (!BOARD_PREFIX) return <BoardList />
  if (error) {
    return (
      <p className="p-6 text-sm text-muted-foreground">
        No board data yet: {error.message}. The first sync may still be running.
      </p>
    )
  }
  if (!data) return <p className="p-6 text-sm text-muted-foreground">Loading the board…</p>

  return (
    <main className="flex min-h-0 flex-1 flex-col">
      <FilterBar
        facets={data.facets}
        columns={data.columns}
        cards={data.cards}
        shown={cards.length}
      />
      {data.readOnly && (
        <p className="px-4 pb-2 text-sm text-muted-foreground">
          This board is read-only: your account can read this repository but not write to it.
        </p>
      )}
      {data.cards.length === 0 ? (
        <EmptyBoard />
      ) : cards.length === 0 ? (
        <NoMatches />
      ) : search.view === 'list' ? (
        <ListView columns={data.columns} cards={cards} />
      ) : (
        <Board columns={data.columns} cards={cards} readOnly={data.readOnly} />
      )}
      <TaskSheet taskIds={taskIds} prLink={data.prLink} />
    </main>
  )
}
