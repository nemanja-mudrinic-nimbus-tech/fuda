import { BOARD_PREFIX, loginPath } from './boardPath'
import type { BoardListData } from './hosts'

export type Column = {
  id: string
  name: string
  statuses: string[]
  prOpen: boolean
  unknown: boolean
}

export type Card = {
  id: string
  title: string
  status: string
  column: string
  owners: string[]
  testers: string[]
  labels: string[]
  prefix: string
  added: string
  claimed: string
  done?: string
  prs: number[]
  prRef?: string
  openPrs: OpenPr[]
  blockedBy?: string
  blockedByIds: string[]
  blocks: string[]
  references: string[]
  referencedBy: string[]
  inProd: boolean
  path: string
}

export type OpenPr = { number: number; url: string }

export type LabelGroup = { name: string; values: string[] }

export type Facets = {
  statuses: string[]
  people: string[]
  testers: string[]
  labelGroups: LabelGroup[]
  labelColors: Record<string, string>
  prefixes: string[]
}

export type Problem = { path: string; reason: string }

export type BranchStatus = { branch: string; sha: string; syncedAt: string; notYet?: boolean }

export type SyncStatus = {
  develop: BranchStatus
  main?: BranchStatus
  lastAttempt: string
  lastError?: string
}

export type Origin = { host: 'github' | 'azure' | 'local'; repo: string; url?: string }

export type BoardData = {
  title: string
  prLink?: string
  origin: Origin
  columns: Column[]
  cards: Card[]
  facets: Facets
  problems: Problem[]
  sync: SyncStatus
  readOnly: boolean
}

export type BoardListing = { host: Origin['host']; repo: string; path: string; title: string }

export type TaskDetail = Card & {
  custom: { key: string; value: string }[]
  html: string
}

export type DocDetail = { path: string; title: string; html: string; backlinks: string[] }

export type GuidePage = { slug: string; title: string }

export type GuideContent = GuidePage & { html: string }

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(
  url: string,
  init?: RequestInit,
  { redirectToLogin = true }: { redirectToLogin?: boolean } = {},
): Promise<T> {
  const res = await fetch(url, init)
  if (res.status === 401 && redirectToLogin) {
    window.location.assign(loginPath(window.location.pathname + window.location.search))
    return new Promise<T>(() => {})
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(res.status, body.error ?? res.statusText)
  }
  return res.status === 202 || res.status === 204 ? (undefined as T) : res.json()
}

const scoped = `/api${BOARD_PREFIX}`

export const api = {
  boards: () => request<BoardListData>('/api/boards'),
  board: () => request<BoardData>(`${scoped}/board`),
  task: (id: string) => request<TaskDetail>(`${scoped}/tasks/${encodeURIComponent(id)}`),
  search: (q: string) => request<string[]>(`${scoped}/search?q=${encodeURIComponent(q)}`),
  archive: () => request<Card[]>(`${scoped}/archive`),
  doc: (path: string) => request<DocDetail>(`${scoped}/docs?path=${encodeURIComponent(path)}`),
  move: (id: string, seen: string, column: string) =>
    request<void>(
      `${scoped}/tasks/${encodeURIComponent(id)}/move`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ seen, column }),
      },
      { redirectToLogin: false },
    ),
  assign: (id: string, seen: string[], owners: string[]) =>
    request<void>(
      `${scoped}/tasks/${encodeURIComponent(id)}/assign`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ seen, owners }),
      },
      { redirectToLogin: false },
    ),
  sync: () => request<void>(`${scoped}/sync`, { method: 'POST' }),
  guide: () => request<GuidePage[]>('/api/guide'),
  guidePage: (slug: string) => request<GuideContent>(`/api/guide/${encodeURIComponent(slug)}`),
}

export function prHref(template: string | undefined, n: number): string | undefined {
  return template?.replace('{n}', String(n))
}
