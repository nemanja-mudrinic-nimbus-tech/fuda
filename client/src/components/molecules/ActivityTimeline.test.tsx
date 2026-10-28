// @vitest-environment jsdom
import { cleanup, fireEvent, render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ActivityTimeline } from './ActivityTimeline'

afterEach(cleanup)

const now = new Date(2026, 9, 27, 11, 12).getTime()

describe('ActivityTimeline', () => {
  it('renders nothing without entries', () => {
    const { container } = render(<ActivityTimeline entries={[]} now={now} />)
    expect(container.innerHTML).toBe('')
  })

  it('shows the newest entry on top, with relative time, badge word and usage, and free text as is', () => {
    render(
      <ActivityTimeline
        now={now}
        entries={[
          {
            when: '2026-10-27 09:12',
            actor: 'factory',
            event: 'spec',
            result: 'READY',
            usage: '12k tokens',
            text: 'first',
          },
          { text: 'a free note' },
          {
            when: '2026-10-27 11:00',
            actor: 'reviewer',
            event: 'code review',
            result: 'VERDICT: BLOCKED',
            text: 'last',
          },
        ]}
      />,
    )
    const items = within(screen.getByRole('list')).getAllByRole('listitem')
    expect(items.map((li) => li.textContent)).toEqual([
      expect.stringContaining('reviewer'),
      'a free note',
      expect.stringContaining('12k tokens'),
    ])
    expect(screen.getByText('BLOCKED').dataset.tone).toBe('fail')
    expect(screen.getByText('READY').dataset.tone).toBe('pass')
    expect(screen.getByText('12 min ago').getAttribute('title')).toBe('2026-10-27 11:00')
    expect(screen.getByText('2 h ago').getAttribute('title')).toBe('2026-10-27 09:12')
  })

  it('shows three detail lines, then the rest on show more', () => {
    render(
      <ActivityTimeline
        now={now}
        entries={[
          {
            when: '2026-10-27 11:00',
            actor: 'reviewer',
            event: 'code review',
            result: 'CHANGES',
            text: 'x',
            details: ['one', 'two', 'three', 'four', 'five'],
          },
        ]}
      />,
    )
    expect(screen.queryByText('four')).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: 'show 2 more' }))
    expect(screen.getByText('five')).toBeTruthy()
    expect(screen.getByRole('button', { name: 'show less' })).toBeTruthy()
  })

  it('links URLs and copies paths', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })
    render(
      <ActivityTimeline
        now={now}
        entries={[
          {
            when: '2026-10-27 11:00',
            actor: 'builder',
            event: 'build',
            result: 'PASS',
            text: 'x',
            details: ['https://example.com/pr/5', '~/runs/F-8/evidence.md'],
          },
        ]}
      />,
    )
    expect(screen.getByRole('link').getAttribute('href')).toBe('https://example.com/pr/5')
    fireEvent.click(screen.getByRole('button', { name: 'Copy ~/runs/F-8/evidence.md' }))
    expect(writeText).toHaveBeenCalledWith('~/runs/F-8/evidence.md')
    expect(await screen.findByRole('button', { name: 'Copied' })).toBeTruthy()
  })
})
