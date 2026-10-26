// @vitest-environment jsdom
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { BoardData, BoardListing } from '@/lib/api'
import type { BoardListData } from '@/lib/hosts'
import { github } from '@/test/hosts'
import { renderWithListing } from '@/test/render'

const hub: BoardListing = {
  host: 'github',
  repo: 'acme/app',
  path: '/github/acme/app',
  title: 'App',
}
const tool: BoardListing = {
  host: 'github',
  repo: 'acme/tool',
  path: '/github/acme/tool',
  title: 'Tool',
}
const notes: BoardListing = { host: 'local', repo: 'notes', path: '/local/notes', title: 'Notes' }

function fakeBoard(listing: BoardListing, url?: string) {
  return {
    title: listing.title,
    origin: { host: listing.host, repo: listing.repo, url },
    sync: { develop: { branch: 'develop' } },
  } as BoardData
}

async function renderPicker(listing: BoardListData, open?: { listed: BoardListing; url?: string }) {
  window.history.replaceState(null, '', open?.listed.path ?? '/')
  vi.resetModules()
  const { WorkspacePicker } = await import('./WorkspacePicker')
  const board = open && fakeBoard(open.listed, open.url)
  renderWithListing(<WorkspacePicker board={board} />, listing)
}

beforeEach(() => window.history.replaceState(null, '', '/'))

describe('WorkspacePicker', () => {
  it('renders no button without Boards or hosts', async () => {
    await renderPicker({ boards: [], hosts: [] })
    expect(screen.queryByRole('button')).toBeNull()
  })

  it('says "Select a Board" when no Board is open', async () => {
    await renderPicker({ boards: [hub], hosts: [github()] })
    expect(screen.getByRole('button', { name: 'Switch Board' }).textContent).toContain(
      'Select a Board',
    )
  })

  it('groups Boards under a label per host, then Local', async () => {
    await renderPicker({ boards: [notes, hub, tool], hosts: [github()] })
    await userEvent.click(screen.getByRole('button', { name: 'Switch Board' }))
    const before = (first: HTMLElement, second: HTMLElement) =>
      Boolean(first.compareDocumentPosition(second) & Node.DOCUMENT_POSITION_FOLLOWING)
    const githubLabel = screen.getByText('GitHub')
    const localLabel = screen.getByText('Local')
    expect(before(githubLabel, screen.getByRole('menuitem', { name: 'App' }))).toBe(true)
    expect(before(screen.getByRole('menuitem', { name: 'Tool' }), localLabel)).toBe(true)
    expect(before(localLabel, screen.getByRole('menuitem', { name: 'Notes' }))).toBe(true)
  })

  it('checks the current Board', async () => {
    await renderPicker({ boards: [hub, tool], hosts: [github()] }, { listed: hub })
    await userEvent.click(screen.getByRole('button', { name: 'Switch Board' }))
    const current = screen.getByRole('menuitem', { name: 'App' })
    const other = screen.getByRole('menuitem', { name: 'Tool' })
    expect(current.getAttribute('aria-current')).toBe('true')
    expect(other.getAttribute('aria-current')).toBeNull()
  })

  it('links "Open repo" in a new tab when the Board has a URL', async () => {
    await renderPicker(
      { boards: [hub], hosts: [github()] },
      { listed: hub, url: 'https://github.com/acme/app' },
    )
    await userEvent.click(screen.getByRole('button', { name: 'Switch Board' }))
    const link = screen.getByRole('menuitem', { name: 'Open repo' })
    expect(link.getAttribute('href')).toBe('https://github.com/acme/app')
    expect(link.getAttribute('target')).toBe('_blank')
  })

  it('hides "Open repo" when the Board has no URL', async () => {
    await renderPicker({ boards: [notes], hosts: [] }, { listed: notes })
    await userEvent.click(screen.getByRole('button', { name: 'Switch Board' }))
    expect(screen.queryByText('Open repo')).toBeNull()
  })
})
