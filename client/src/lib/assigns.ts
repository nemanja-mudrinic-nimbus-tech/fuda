import type { BoardData } from './api'

export type PendingAssign = { cardId: string; owners: string[] }

export function applyPendingAssigns(board: BoardData, assigns: PendingAssign[]): BoardData {
  if (assigns.length === 0) return board
  const target = new Map(assigns.map((a) => [a.cardId, a.owners]))
  return {
    ...board,
    cards: board.cards.map((card) => {
      const owners = target.get(card.id)
      return owners ? { ...card, owners } : card
    }),
  }
}

export function toggleOwner(owners: string[], name: string): string[] {
  return owners.includes(name) ? owners.filter((o) => o !== name) : [...owners, name]
}

export function ownerChoices(people: string[], owners: string[]): string[] {
  return [...people, ...owners.filter((o) => !people.includes(o))]
}
