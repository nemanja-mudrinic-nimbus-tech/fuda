import {
  keepPreviousData,
  useMutation,
  useMutationState,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import { useMemo } from 'react'
import { toast } from 'sonner'
import { ApiError, api, type Card } from './api'
import { BOARD_PREFIX, loginPath } from './boardPath'
import { applyPendingAssigns, type PendingAssign } from './assigns'
import { applyPendingMoves, type PendingMove } from './moves'

export const keys = {
  boards: ['boards'] as const,
  board: ['board'] as const,
  task: (id: string) => ['task', id] as const,
  search: (q: string) => ['search', q] as const,
  archive: ['archive'] as const,
  doc: (path: string) => ['doc', path] as const,
  guide: ['guide'] as const,
  guidePage: (slug: string) => ['guide', slug] as const,
}

export function useBoardListing() {
  return useQuery({ queryKey: keys.boards, queryFn: api.boards, staleTime: 5 * 60_000 })
}

export function useBoard() {
  const query = useQuery({
    queryKey: keys.board,
    queryFn: api.board,
    refetchInterval: 5_000,
    staleTime: 0,
    enabled: !!BOARD_PREFIX,
  })
  const moves = usePendingMoves()
  const assigns = usePendingAssigns()
  const data = useMemo(
    () =>
      query.data ? applyPendingAssigns(applyPendingMoves(query.data, moves), assigns) : undefined,
    [query.data, moves, assigns],
  )
  return { ...query, data }
}

const moveKey = ['move'] as const

type MoveVariables = { card: Card; column: string }

export function usePendingMoves(): PendingMove[] {
  const saving = useMutationState({
    filters: { mutationKey: moveKey, status: 'pending' },
    select: (mutation) => mutation.state.variables as MoveVariables,
  })
  return useMemo(() => saving.map((v) => ({ cardId: v.card.id, column: v.column })), [saving])
}

function reportSaveFailure(error: unknown, what: string, retry: string) {
  if (error instanceof ApiError && error.status === 401) {
    toast.error('Your login expired', {
      description: `The card went back. Log in again, then ${retry}.`,
      action: {
        label: 'Log in again',
        onClick: () =>
          window.location.assign(loginPath(window.location.pathname + window.location.search)),
      },
    })
    return
  }
  toast.error(error instanceof ApiError ? error.message : `The ${what} was not saved`, {
    description: 'The card went back.',
  })
}

export function useMove() {
  const client = useQueryClient()
  return useMutation({
    mutationKey: moveKey,
    mutationFn: ({ card, column }: MoveVariables) => api.move(card.id, card.status, column),
    onSettled: () => client.invalidateQueries({ queryKey: keys.board }),
    onError: (error) => reportSaveFailure(error, 'move', 'move it'),
  })
}

const assignKey = ['assign'] as const

type AssignVariables = { card: Pick<Card, 'id' | 'owners'>; owners: string[] }

export function usePendingAssigns(): PendingAssign[] {
  const saving = useMutationState({
    filters: { mutationKey: assignKey, status: 'pending' },
    select: (mutation) => mutation.state.variables as AssignVariables,
  })
  return useMemo(() => saving.map((v) => ({ cardId: v.card.id, owners: v.owners })), [saving])
}

export function useAssign() {
  const client = useQueryClient()
  return useMutation({
    mutationKey: assignKey,
    mutationFn: ({ card, owners }: AssignVariables) => api.assign(card.id, card.owners, owners),
    onSettled: (_data, _error, { card }) =>
      Promise.all([
        client.invalidateQueries({ queryKey: keys.board }),
        client.invalidateQueries({ queryKey: keys.task(card.id) }),
      ]),
    onError: (error) => reportSaveFailure(error, 'change of owners', 'pick the owners again'),
  })
}

export function useTask(id: string | undefined) {
  return useQuery({
    queryKey: keys.task(id ?? ''),
    queryFn: () => api.task(id!),
    enabled: !!id,
  })
}

export function useTextSearch(q: string | undefined, enabled: boolean) {
  const query = q?.trim() ?? ''
  return useQuery({
    queryKey: keys.search(query),
    queryFn: () => api.search(query),
    enabled: enabled && query.length > 1,
    placeholderData: keepPreviousData,
  })
}

export function useArchive() {
  return useQuery({ queryKey: keys.archive, queryFn: api.archive })
}

export function useDoc(path: string) {
  return useQuery({ queryKey: keys.doc(path), queryFn: () => api.doc(path) })
}

export function useSync() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: api.sync,
    onSuccess: () => {
      toast.success('Reading the repo…')
      setTimeout(() => void client.invalidateQueries(), 1500)
    },
    onError: (error) => {
      if (error instanceof ApiError && error.status === 429) {
        toast.info('Synced moments ago', { description: 'Try again in a few seconds.' })
        return
      }
      toast.error('Sync failed', { description: error.message })
    },
  })
}

export function useGuide() {
  return useQuery({ queryKey: keys.guide, queryFn: api.guide, staleTime: Infinity })
}

export function useGuidePage(slug: string) {
  return useQuery({
    queryKey: keys.guidePage(slug),
    queryFn: () => api.guidePage(slug),
    staleTime: Infinity,
  })
}
