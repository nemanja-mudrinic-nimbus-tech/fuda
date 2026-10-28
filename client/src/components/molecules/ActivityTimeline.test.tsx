// @vitest-environment jsdom
import { cleanup, render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { ActivityTimeline } from './ActivityTimeline'

afterEach(cleanup)

describe('ActivityTimeline', () => {
  it('renders nothing without entries', () => {
    const { container } = render(<ActivityTimeline entries={[]} />)
    expect(container.innerHTML).toBe('')
  })

  it('shows the newest entry on top, with badge and usage, and free text as is', () => {
    render(
      <ActivityTimeline
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
            result: 'BLOCKED',
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
  })
})
