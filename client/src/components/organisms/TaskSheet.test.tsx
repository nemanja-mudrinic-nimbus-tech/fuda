// @vitest-environment jsdom
import { act, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { TaskSheet } from './TaskSheet'

vi.mock('@/hooks/useBoardSearch', () => ({
  useBoardSearch: () => ({
    search: { task: 'FUDA-12' },
    toggle: vi.fn(),
    openTask: vi.fn(),
  }),
}))

vi.mock('@/hooks/useLeaveWarning', () => ({ useLeaveWarning: vi.fn() }))

vi.mock('@/lib/queries', () => ({
  useTask: () => ({ data: undefined, isError: false, dataUpdatedAt: 0 }),
  useBoard: () => ({ data: { cards: [], columns: [], readOnly: true }, dataUpdatedAt: 0 }),
  useAssign: () => ({}),
  useMove: () => ({}),
  usePendingMoves: () => [],
  usePendingAssigns: () => [],
}))

describe('TaskSheet copy task ID', () => {
  const writeText = vi.fn(() => Promise.resolve())

  beforeEach(() => {
    vi.useFakeTimers()
    writeText.mockClear()
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('copies just the ID on a read-only board and shows Copied for 1.5s', async () => {
    render(<TaskSheet taskIds={['FUDA-12']} />)
    const button = screen.getByRole('button', { name: 'Copy task ID' })
    expect(button.textContent).toBe('')

    await act(async () => {
      fireEvent.click(button)
    })
    expect(writeText).toHaveBeenCalledWith('FUDA-12')
    expect(button.textContent).toBe('Copied')

    act(() => {
      vi.advanceTimersByTime(1500)
    })
    expect(button.textContent).toBe('')
  })
})
