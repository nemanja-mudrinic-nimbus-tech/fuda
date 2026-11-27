import type { BoardData, Card, Column } from './api'

export type PendingMove = { cardId: string; column: string }

export function applyPendingMoves(board: BoardData, moves: PendingMove[]): BoardData {
  if (moves.length === 0) {
    return board
  }

  const target = new Map(moves.map((m) => [m.cardId, m.column]))
  const columns = new Map(board.columns.map((c) => [c.id, c]))

  return {
    ...board,
    cards: board.cards.map((card) => {
      const column = columns.get(target.get(card.id) ?? '')

      if (!column) {
        return card
      }

      return { ...card, column: column.id, status: column.statuses[0] ?? card.status }
    }),
  }
}

export function canDrag(card: Card, saving: boolean, readOnly: boolean): boolean {
  return !readOnly && !saving && !card.archived && card.openPrs.length === 0
}

export function columnStatus(column: Column): string {
  if (column.prOpen) {
    return 'in review'
  }

  if (column.done) {
    return 'done'
  }

  return column.statuses[0] ?? ''
}

// The status a card is drawn with: the Stage the board put it in wins over the
// file's `status` (a task with an open PR sits in In review whatever its status).
export function cardStatus(card: Card, columns: Column[]): string {
  const column = columns.find((c) => c.id === card.column)

  return column?.prOpen ? columnStatus(column) : card.status
}

export function canDrop(card: Card, column: Column): boolean {
  return column.statuses.length > 0 && card.column !== column.id
}

export function moveTargets(
  card: Card,
  columns: Column[],
  saving: boolean,
  readOnly: boolean,
): Column[] {
  return canDrag(card, saving, readOnly) ? columns.filter((column) => canDrop(card, column)) : []
}
