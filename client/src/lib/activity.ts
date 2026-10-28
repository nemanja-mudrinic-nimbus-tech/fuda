export type ResultTone = 'pass' | 'warn' | 'fail' | 'other'

const tones: Record<string, ResultTone> = {
  PASS: 'pass',
  READY: 'pass',
  COMPLETE: 'pass',
  auto_merge: 'pass',
  CHANGES: 'warn',
  GAPS: 'warn',
  needs_human: 'warn',
  BLOCKED: 'fail',
  FAILED: 'fail',
  HIT: 'fail',
}

export function resultTone(result: string): ResultTone {
  return Object.hasOwn(tones, result) ? tones[result] : 'other'
}
