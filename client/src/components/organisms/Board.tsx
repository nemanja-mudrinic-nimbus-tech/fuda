import { useState } from 'react'
import { Column } from '@/components/organisms/Column'
import { useCollapsedColumns } from '@/hooks/useCollapsedColumns'
import { useLeaveWarning } from '@/hooks/useLeaveWarning'
import type { Card, Column as ColumnData } from '@/lib/api'
import { canDrag, canDrop } from '@/lib/moves'
import { useMove, usePendingMoves } from '@/lib/queries'

export function Board({
  columns,
  cards,
  readOnly,
}: {
  columns: ColumnData[]
  cards: Card[]
  readOnly: boolean
}) {
  const { isCollapsed, toggle } = useCollapsedColumns()
  const [dragging, setDragging] = useState<Card | null>(null)
  const move = useMove()
  const saving = new Set(usePendingMoves().map((m) => m.cardId))
  useLeaveWarning(saving.size > 0)
  return (
    <div className="flex min-h-0 flex-1 snap-x snap-mandatory gap-3 overflow-x-auto px-3 pb-3 sm:snap-none sm:px-4 sm:pb-4">
      {columns.map((column) => (
        <Column
          key={column.id}
          column={column}
          cards={cards.filter((c) => c.column === column.id)}
          saving={saving}
          dropAllowed={dragging !== null && canDrop(dragging, column)}
          isDraggable={(card) => canDrag(card, saving.has(card.id), readOnly)}
          onDragStart={setDragging}
          onDragEnd={() => setDragging(null)}
          onDrop={() => {
            if (dragging) move.mutate({ card: dragging, column: column.id })
            setDragging(null)
          }}
          collapsed={isCollapsed(column.id)}
          onToggleCollapse={() => toggle(column.id)}
        />
      ))}
    </div>
  )
}
