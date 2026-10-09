import { describe, expect, it } from 'vitest'
import { groupBoards, onlyBoard, pickerButtonLabel, showLogOutAll } from './hosts'

const github = { host: 'github', loggedIn: true, login: '/auth/github/login' } as const
const azure = { host: 'azure', loggedIn: false, login: '/auth/azure/login' } as const
const board = {
  host: 'github',
  repo: 'o/fuda-x',
  path: '/github/o/fuda-x',
  title: 'fuda-x',
} as const

describe('onlyBoard', () => {
  it('opens a single Board when every host is logged in', () => {
    expect(onlyBoard({ boards: [board], hosts: [github] })).toEqual(board)
  })
  it('keeps the list when another host still needs a login', () => {
    expect(onlyBoard({ boards: [board], hosts: [github, azure] })).toBeUndefined()
  })
})

describe('groupBoards', () => {
  const azureBoard = {
    ...board,
    host: 'azure',
    path: '/azure/o/p/fuda-y',
    title: 'fuda-y',
  } as const
  const localBoard = { ...board, host: 'local', path: '/local/z', title: 'z' } as const

  it('groups Boards by host, then Local last', () => {
    const groups = groupBoards([localBoard, azureBoard, board])
    expect(groups.map((group) => group.label)).toEqual(['GitHub', 'Azure DevOps', 'Local'])
    expect(groups.map((group) => group.boards)).toEqual([[board], [azureBoard], [localBoard]])
  })
  it('leaves out a host with no Boards', () => {
    expect(groupBoards([board]).map((group) => group.label)).toEqual(['GitHub'])
  })
  it('returns no groups for no Boards', () => {
    expect(groupBoards([])).toEqual([])
  })
})

describe('pickerButtonLabel', () => {
  it('shows the Board name', () => {
    expect(pickerButtonLabel({ title: 'Mine', origin: { repo: 'o/fuda-x' } })).toBe('Mine')
  })
  it('falls back to the repository', () => {
    expect(pickerButtonLabel({ title: '', origin: { repo: 'o/fuda-x' } })).toBe('o/fuda-x')
  })
  it('uses the listed Board while the Board itself loads', () => {
    expect(pickerButtonLabel(undefined, { title: 'Listed' })).toBe('Listed')
  })
  it('asks to select a Board when none is open', () => {
    expect(pickerButtonLabel(undefined)).toBe('Select a Board')
  })
})

describe('showLogOutAll', () => {
  it('needs more than one host logged in', () => {
    expect(showLogOutAll([github, azure])).toBe(false)
    expect(showLogOutAll([github, { ...azure, loggedIn: true }])).toBe(true)
    expect(showLogOutAll([github])).toBe(false)
  })
})
