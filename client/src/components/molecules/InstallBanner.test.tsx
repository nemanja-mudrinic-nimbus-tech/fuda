// @vitest-environment jsdom
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import type { BoardListData } from '@/lib/hosts'
import { github } from '@/test/hosts'
import { renderWithListing } from '@/test/render'
import { InstallBanner } from './InstallBanner'

const url = 'https://github.com/apps/fuda/installations/new'

function listing(appInstalled?: boolean): BoardListData {
  return { boards: [], hosts: [{ ...github(), installUrl: url, appInstalled }] }
}

describe('InstallBanner', () => {
  it('shows the install link when the App is not installed', () => {
    renderWithListing(<InstallBanner />, listing(false))
    expect(
      screen.getByRole('link', { name: 'Install the fuda GitHub App' }).getAttribute('href'),
    ).toBe(url)
  })

  it.each([
    ['installed', true],
    ['unknown', undefined],
  ])('shows nothing when the install state is %s', (_name, appInstalled) => {
    renderWithListing(<InstallBanner />, listing(appInstalled))
    expect(screen.queryByRole('status')).toBeNull()
  })

  it('can be closed', async () => {
    renderWithListing(<InstallBanner />, listing(false))
    await userEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(screen.queryByRole('status')).toBeNull()
  })
})
