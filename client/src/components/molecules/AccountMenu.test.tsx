// @vitest-environment jsdom
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import { azure, github } from '@/test/hosts'
import { renderWithListing } from '@/test/render'
import type { HostStatus } from '@/lib/hosts'
import { AccountMenu } from './AccountMenu'

async function openMenu(hosts: HostStatus[]) {
  renderWithListing(<AccountMenu />, { boards: [], hosts })
  await userEvent.click(screen.getByRole('button', { name: 'Account' }))
}

describe('AccountMenu', () => {
  it('renders no button without hosts', () => {
    renderWithListing(<AccountMenu />, { boards: [], hosts: [] })
    expect(screen.queryByRole('button', { name: 'Account' })).toBeNull()
  })

  it('shows each host with its own state and action, reachable as menu items', async () => {
    await openMenu([github(true), azure(false)])
    const githubItem = screen.getByRole('menuitem', { name: /GitHub/ })
    const azureItem = screen.getByRole('menuitem', { name: /Azure DevOps/ })
    expect(within(githubItem).getByText('Logged in')).toBeTruthy()
    expect(within(githubItem).getByText('Log out')).toBeTruthy()
    expect(within(azureItem).getByText('Logged out')).toBeTruthy()
    expect(within(azureItem).getByText('Log in')).toBeTruthy()
    expect(azureItem.getAttribute('href')).toMatch(/^\/auth\/azure\/login\?return=/)
  })

  it.each([
    ['one host is logged in', [github(true), azure(false)], false],
    ['no host is logged in', [github(false), azure(false)], false],
    ['two hosts are logged in', [github(true), azure(true)], true],
  ])('Log out of all when %s: shown is %s', async (_name, hosts, shown) => {
    await openMenu(hosts)
    expect(screen.queryByText('Log out of all') !== null).toBe(shown)
  })
})
