import { describe, expect, it } from 'vitest'
import type { BoardData, Card } from './api'
import { applyPendingAssigns, ownerChoices, toggleOwner } from './assigns'

const card = (id: string, owners: string[]): Card => ({ id, owners }) as Card
const board = { cards: [card('A-1', ['Ann']), card('A-2', [])] } as BoardData

describe('applyPendingAssigns', () => {
  it('shows the owners that are saving', () => {
    const shown = applyPendingAssigns(board, [{ cardId: 'A-1', owners: ['Ben'] }])
    expect(shown.cards[0].owners).toEqual(['Ben'])
    expect(shown.cards[1]).toBe(board.cards[1])
  })
  it('shows an empty list when every owner is removed', () => {
    expect(applyPendingAssigns(board, [{ cardId: 'A-1', owners: [] }]).cards[0].owners).toEqual([])
  })
  it('returns the board itself when nothing is saving', () => {
    expect(applyPendingAssigns(board, [])).toBe(board)
  })
})

describe('toggleOwner', () => {
  it('adds a name that is not an owner', () => {
    expect(toggleOwner(['Ann'], 'Ben')).toEqual(['Ann', 'Ben'])
  })
  it('removes a name that is an owner', () => {
    expect(toggleOwner(['Ann', 'Ben'], 'Ann')).toEqual(['Ben'])
  })
})

describe('ownerChoices', () => {
  it('lists the people, then owners who are not in the list', () => {
    expect(ownerChoices(['Ann', 'Ben'], ['Ben', 'Cy'])).toEqual(['Ann', 'Ben', 'Cy'])
  })
})
