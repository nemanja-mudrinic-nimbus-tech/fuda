import type { BoardListing } from './api'

export type HostName = 'github' | 'azure'

export type HostStatus = { host: HostName; loggedIn: boolean; login: string; error?: string }

export type BoardListData = { boards: BoardListing[]; hosts: HostStatus[] }

const labels: Record<HostName, string> = { github: 'GitHub', azure: 'Azure DevOps' }

export function hostLabel(host: HostName): string {
  return labels[host]
}

export function onlyBoard({ boards, hosts }: BoardListData): BoardListing | undefined {
  return boards.length === 1 && hosts.every((host) => host.loggedIn) ? boards[0] : undefined
}

export async function logOut(host?: HostName) {
  await fetch(host ? `/auth/${host}/logout` : '/auth/logout', { method: 'POST' })
  window.location.assign('/')
}

const groupOrder = ['github', 'azure', 'local'] as const

export function groupBoards(boards: BoardListing[]) {
  return groupOrder
    .map((host) => ({
      host,
      label: host === 'local' ? 'Local' : hostLabel(host),
      boards: boards.filter((board) => board.host === host),
    }))
    .filter((group) => group.boards.length > 0)
}

export function pickerButtonLabel(
  board?: { title: string; origin: { repo: string } },
  listed?: { title: string },
): string {
  if (board) return board.title || board.origin.repo
  return listed?.title ?? 'Select a Board'
}

export function showLogOutAll(hosts: HostStatus[]): boolean {
  return hosts.filter((host) => host.loggedIn).length > 1
}
