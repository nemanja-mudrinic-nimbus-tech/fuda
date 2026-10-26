import { BOARD_PREFIX } from '@/lib/boardPath'
import { useRouter } from '@tanstack/react-router'
import { useEffect, useMemo, useRef } from 'react'
import { cn } from '@/lib/utils'

function escapeRegExp(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function idPattern(ids: string[]): RegExp | undefined {
  if (ids.length === 0) return undefined
  const alternatives = [...ids].sort((a, b) => b.length - a.length).map(escapeRegExp)
  return new RegExp(`(?<![A-Za-z0-9-])(${alternatives.join('|')})(?![A-Za-z0-9])`, 'g')
}

function linkTaskIds(root: HTMLElement, pattern: RegExp, selfId?: string) {
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
    acceptNode: (node) =>
      node.parentElement?.closest('a, code, pre')
        ? NodeFilter.FILTER_REJECT
        : NodeFilter.FILTER_ACCEPT,
  })
  const texts: Text[] = []
  while (walker.nextNode()) texts.push(walker.currentNode as Text)

  for (const text of texts) {
    const value = text.nodeValue ?? ''
    pattern.lastIndex = 0
    if (!pattern.test(value)) continue
    pattern.lastIndex = 0
    const fragment = document.createDocumentFragment()
    let last = 0
    for (const match of value.matchAll(pattern)) {
      const id = match[0]
      fragment.append(value.slice(last, match.index))
      if (id === selfId) {
        fragment.append(id)
      } else {
        const a = document.createElement('a')
        a.href = `/?task=${encodeURIComponent(id)}`
        a.textContent = id
        a.className = 'task-ref'
        fragment.append(a)
      }
      last = match.index + id.length
    }
    fragment.append(value.slice(last))
    text.replaceWith(fragment)
  }
}

function addCopyButtons(root: HTMLElement) {
  for (const pre of root.querySelectorAll('pre')) {
    const code = pre.querySelector('code')
    if (!code || code.classList.contains('language-mermaid')) continue
    const button = document.createElement('button')
    button.type = 'button'
    button.className = 'copy-code'
    button.textContent = 'Copy'
    button.addEventListener('click', () => {
      void navigator.clipboard.writeText(code.textContent ?? '').then(() => {
        button.textContent = 'Copied'
        setTimeout(() => (button.textContent = 'Copy'), 1500)
      })
    })
    pre.classList.add('has-copy')
    pre.append(button)
  }
}

async function renderMermaid(root: HTMLElement) {
  const blocks = [...root.querySelectorAll<HTMLElement>('code.language-mermaid')]
  if (blocks.length === 0) return
  const { default: mermaid } = await import('mermaid')
  mermaid.initialize({
    startOnLoad: false,
    theme: document.documentElement.classList.contains('dark') ? 'dark' : 'neutral',
    securityLevel: 'strict',
  })
  for (const [i, code] of blocks.entries()) {
    const pre = code.closest('pre') ?? code
    try {
      const { svg } = await mermaid.render(`mermaid-${Date.now()}-${i}`, code.textContent ?? '')
      const figure = document.createElement('div')
      figure.className = 'mermaid-diagram'
      figure.innerHTML = svg
      pre.replaceWith(figure)
    } catch {
      pre.classList.add('mermaid-error')
    }
  }
}

function promoteSection(root: HTMLElement, heading: string) {
  const h2 = [...root.querySelectorAll('h2')].find(
    (h) => h.textContent?.trim().toLowerCase() === heading.toLowerCase(),
  )
  if (!h2 || h2 === root.firstElementChild) return
  const section: Element[] = [h2]
  for (let el = h2.nextElementSibling; el && el.tagName !== 'H2'; el = el.nextElementSibling)
    section.push(el)
  const box = document.createElement('div')
  box.className = 'lead-section'
  box.append(...section)
  root.prepend(box)
}

export function RichContent({
  html,
  taskIds,
  selfId,
  lead,
  dropTitle,
  className,
}: {
  html: string
  taskIds: string[]
  selfId?: string
  lead?: string
  dropTitle?: boolean
  className?: string
}) {
  const ref = useRef<HTMLDivElement>(null)
  const router = useRouter()
  const pattern = useMemo(() => idPattern(taskIds), [taskIds])

  useEffect(() => {
    const root = ref.current
    if (!root) return
    root.innerHTML = html
    if (dropTitle) root.querySelector('h1')?.remove()
    if (lead) promoteSection(root, lead)
    if (pattern) linkTaskIds(root, pattern, selfId)
    addCopyButtons(root)
    void renderMermaid(root)
  }, [html, pattern, selfId, lead, dropTitle])

  return (
    <div
      ref={ref}
      className={cn(
        'prose prose-sm max-w-none dark:prose-invert prose-headings:scroll-mt-4 prose-h1:text-2xl prose-h2:text-lg prose-a:text-primary prose-code:rounded prose-code:bg-muted prose-code:px-1 prose-code:py-0.5 prose-code:font-normal prose-code:[overflow-wrap:anywhere] prose-code:before:content-none prose-code:after:content-none prose-pre:bg-muted prose-pre:text-foreground prose-img:rounded-lg prose-img:border prose-img:border-border prose-img:bg-card',
        className,
      )}
      onClick={(e) => {
        const anchor = (e.target as HTMLElement).closest('a')
        const href = anchor?.getAttribute('href')
        if (!anchor || !href || anchor.target === '_blank' || !href.startsWith('/')) return
        e.preventDefault()
        void router.navigate({
          href: href.startsWith(BOARD_PREFIX + '/') || !BOARD_PREFIX ? href : BOARD_PREFIX + href,
        })
      }}
    />
  )
}
