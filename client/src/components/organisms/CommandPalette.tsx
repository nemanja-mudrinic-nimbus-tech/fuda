import { useNavigate } from '@tanstack/react-router'
import {
  Archive,
  BookOpen,
  ChartColumn,
  FileText,
  FilterX,
  LayoutGrid,
  Moon,
  RefreshCw,
  Search,
  Sun,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { StatusIcon } from '@/components/atoms/StatusIcon'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from '@/components/ui/command'
import { Kbd } from '@/components/ui/kbd'
import { useTheme } from '@/hooks/useTheme'
import { useBoard, useGuide, useSync } from '@/lib/queries'
import { cardStatus } from '@/lib/moves'

export function CommandPalette() {
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()
  const { data: board } = useBoard()
  const { data: guide = [] } = useGuide()
  const sync = useSync()
  const { resolved, setChoice } = useTheme()

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key.toLowerCase() === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((v) => !v)
      }
    }
    window.addEventListener('keydown', onKey)

    return () => window.removeEventListener('keydown', onKey)
  }, [])

  const run = (action: () => void) => {
    setOpen(false)
    action()
  }

  return (
    <>
      <Button
        variant="outline"
        size="sm"
        className="h-7 gap-2 bg-card px-2 text-xs text-muted-foreground shadow-xs"
        onClick={() => setOpen(true)}
        aria-label="Open command palette"
      >
        <Search className="size-3.5" />
        <span className="hidden md:inline">Jump to…</span>
        <Kbd className="hidden md:inline-flex">⌘K</Kbd>
      </Button>
      <CommandDialog
        open={open}
        onOpenChange={setOpen}
        title="Jump to"
        className="sm:max-w-xl"
        description="Find a task, page or action"
      >
        <Command>
          <CommandInput placeholder="Type a task id, title, page or action…" />
          <CommandList className="max-h-[min(28rem,60vh)]">
            <CommandEmpty>Nothing found.</CommandEmpty>
            <CommandGroup heading="Go to">
              <CommandItem onSelect={() => run(() => void navigate({ to: '/' }))}>
                <LayoutGrid />
                Board
              </CommandItem>
              <CommandItem onSelect={() => run(() => void navigate({ to: '/insights' }))}>
                <ChartColumn />
                Insights
              </CommandItem>
              <CommandItem onSelect={() => run(() => void navigate({ to: '/archive' }))}>
                <Archive />
                Archive
              </CommandItem>
              <CommandItem
                onSelect={() => run(() => void navigate({ to: '/docs/$', params: { _splat: '' } }))}
              >
                <FileText />
                Docs
              </CommandItem>
              {guide.map((page) => (
                <CommandItem
                  key={page.slug}
                  value={`guide ${page.title}`}
                  onSelect={() =>
                    run(() => void navigate({ to: '/guide/$', params: { _splat: page.slug } }))
                  }
                >
                  <BookOpen />
                  Guide: {page.title}
                </CommandItem>
              ))}
            </CommandGroup>
            <CommandSeparator />
            <CommandGroup heading="Actions">
              <CommandItem onSelect={() => run(() => sync.mutate())}>
                <RefreshCw />
                Sync now
              </CommandItem>
              <CommandItem
                onSelect={() => run(() => setChoice(resolved === 'dark' ? 'light' : 'dark'))}
              >
                {resolved === 'dark' ? <Sun /> : <Moon />}
                Switch to {resolved === 'dark' ? 'light' : 'dark'} theme
              </CommandItem>
              <CommandItem onSelect={() => run(() => void navigate({ to: '/', search: {} }))}>
                <FilterX />
                Clear board filters
              </CommandItem>
            </CommandGroup>
            {board && (
              <>
                <CommandSeparator />
                <CommandGroup heading="Tasks">
                  {board.cards.map((card) => (
                    <CommandItem
                      key={card.id}
                      value={`${card.id} ${card.title}`}
                      onSelect={() =>
                        run(
                          () =>
                            void navigate({
                              to: '/',
                              search: (prev) => ({ ...prev, task: card.id }),
                            }),
                        )
                      }
                    >
                      <StatusIcon status={cardStatus(card, board.columns)} className="size-3" />
                      <span className="w-16 shrink-0 font-mono text-xs text-muted-foreground">
                        {card.id}
                      </span>
                      <span className="truncate">{card.title}</span>
                      <CommandShortcut>{cardStatus(card, board.columns)}</CommandShortcut>
                    </CommandItem>
                  ))}
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </CommandDialog>
    </>
  )
}
