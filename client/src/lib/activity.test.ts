import { describe, expect, it } from 'vitest'
import { resultTone } from './activity'

describe('resultTone', () => {
  it.each([
    ['PASS', 'pass'],
    ['READY', 'pass'],
    ['COMPLETE', 'pass'],
    ['auto_merge', 'pass'],
    ['CHANGES', 'warn'],
    ['GAPS', 'warn'],
    ['needs_human', 'warn'],
    ['BLOCKED', 'fail'],
    ['FAILED', 'fail'],
    ['HIT', 'fail'],
    ['pass', 'other'],
    ['SKIPPED', 'other'],
    ['toString', 'other'],
  ])('%s is %s', (result, tone) => {
    expect(resultTone(result)).toBe(tone)
  })
})
