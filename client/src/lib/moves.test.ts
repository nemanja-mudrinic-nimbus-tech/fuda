import { describe, expect, it } from 'vitest'
import type { BoardData, Card, Column } from './api'
import { applyPendingMoves, canDrag, canDrop } from './moves'

const column = (id: string, extra: Partial<Column> = {}): Column => ({
  id,
  name: id,
  statuses: [id.replace('-', ' ')],
  prOpen: false,
  unknown: false,
  ...extra,
})

const card = (id: string, extra: Partial<Card> = {}): Card =>
  ({ id, status: 'backlog', column: 'backlog', openPrs: [], ...extra }) as Card

const board = {
  columns: [
    column('backlog'),
    column('in-progress'),
    column('in-review', { prOpen: true, statuses: [] }),
  ],
  cards: [card('A-1'), card('A-2')],
} as BoardData

describe('applyPendingMoves', () => {
  it('shows a saving card in its new column and status', () => {
    const shown = applyPendingMoves(board, [{ cardId: 'A-1', column: 'in-progress' }])
    expect(shown.cards[0]).toMatchObject({ column: 'in-progress', status: 'in progress' })
    expect(shown.cards[1]).toBe(board.cards[1])
  })
  it('returns the board itself when nothing is saving', () => {
    expect(applyPendingMoves(board, [])).toBe(board)
  })
  it('ignores a move to a column that is gone', () => {
    expect(applyPendingMoves(board, [{ cardId: 'A-1', column: 'nowhere' }]).cards[0]).toBe(
      board.cards[0],
    )
  })
})

describe('canDrag', () => {
  it('refuses a card that is saving', () => {
    expect(canDrag(card('A-1'), true, false)).toBe(false)
  })
  it('refuses a card locked by an open PR', () => {
    expect(canDrag(card('A-1', { openPrs: [{ number: 7, url: 'u' }] }), false, false)).toBe(false)
  })
  it('refuses every card on a read-only board', () => {
    expect(canDrag(card('A-1'), false, true)).toBe(false)
  })
  it('allows a free card', () => {
    expect(canDrag(card('A-1'), false, false)).toBe(true)
  })
})

describe('canDrop', () => {
  it('refuses its own column and PR-derived columns', () => {
    expect(canDrop(card('A-1'), board.columns[0]!)).toBe(false)
    expect(canDrop(card('A-1'), board.columns[2]!)).toBe(false)
  })
  it('accepts another column', () => {
    expect(canDrop(card('A-1'), board.columns[1]!)).toBe(true)
  })
})
