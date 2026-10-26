import type { HostStatus } from '@/lib/hosts'

export const github = (loggedIn = true): HostStatus => ({
  host: 'github',
  loggedIn,
  login: '/auth/github/login',
})

export const azure = (loggedIn = true): HostStatus => ({
  host: 'azure',
  loggedIn,
  login: '/auth/azure/login',
})
