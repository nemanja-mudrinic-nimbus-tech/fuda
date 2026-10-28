// @vitest-environment jsdom
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { Column } from '@/lib/api'
import { StagePicker } from './StagePicker'

afterEach(cleanup)

const column = (id: string, name: string): Column => ({
  id,
  name,
  statuses: [name.toLowerCase()],
  prOpen: false,
  done: false,
  unknown: false,
  activity: false,
})

describe('StagePicker', () => {
  it('renders nothing when the card cannot move', () => {
    render(<StagePicker targets={[]} saving={false} onMove={() => {}} />)
    expect(screen.queryByRole('button', { name: 'Move to' })).toBeNull()
  })

  it('moves the card to the chosen stage with the keyboard', async () => {
    const onMove = vi.fn()
    const testing = column('testing', 'Testing')
    render(
      <StagePicker
        targets={[column('in-progress', 'In progress'), testing]}
        saving={false}
        onMove={onMove}
      />,
    )
    screen.getByRole('button', { name: 'Move to' }).focus()
    await userEvent.keyboard('{Enter}')
    await userEvent.keyboard('{ArrowDown}{ArrowDown}{Enter}')
    expect(onMove).toHaveBeenCalledWith(testing)
  })

  it('is disabled while saving', () => {
    render(<StagePicker targets={[column('testing', 'Testing')]} saving onMove={() => {}} />)
    expect(screen.getByRole('button', { name: 'Saving…' }).hasAttribute('disabled')).toBe(true)
  })
})
