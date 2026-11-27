import { useQueryClient } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import { Loader2Icon } from 'lucide-react'
import { lazy, Suspense, useState } from 'react'
import { toast } from 'sonner'
import { RichContent } from '@/components/molecules/RichContent'
import { TaskLink } from '@/components/atoms/TaskLink'
import { DocTree } from '@/components/organisms/DocTree'
import { Button } from '@/components/ui/button'
import { ApiError, type DocSource } from '@/lib/api'
import { docSourceQuery, useBoard, useDoc, useDocList } from '@/lib/queries'

const loadEditor = () => import('@/components/organisms/DocEditor')
const DocEditor = lazy(loadEditor)

export const Route = createFileRoute('/docs/$')({
  component: DocPage,
})

const defaultDoc = 'board/README.md'

function DocPage() {
  const { _splat: requested = '' } = Route.useParams()
  const { data: paths } = useDocList()
  const { data: board } = useBoard()
  const client = useQueryClient()
  const path = requested || (paths?.includes(defaultDoc) ? defaultDoc : '')
  const { data: doc, error, isPending } = useDoc(path)
  const [editing, setEditing] = useState<{ path: string; source: DocSource } | null>(null)
  const [loadingSource, setLoadingSource] = useState(false)
  if (editing && editing.path !== path) setEditing(null)
  const editor = editing?.path === path ? editing.source : undefined

  const warmEditor = () => {
    void client.prefetchQuery(docSourceQuery(path))
    loadEditor().catch(() => undefined)
  }

  const startEditing = async () => {
    setLoadingSource(true)
    try {
      const [source] = await Promise.all([client.fetchQuery(docSourceQuery(path)), loadEditor()])
      setEditing({ path, source })
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : 'The doc could not be opened for editing')
    } finally {
      setLoadingSource(false)
    }
  }

  return (
    <div className="flex min-h-0 flex-1">
      <DocTree paths={paths ?? []} openPath={path} />
      <main className={editor ? 'flex flex-1 flex-col overflow-hidden' : 'flex-1 overflow-y-auto'}>
        <div className={editor ? 'flex min-h-0 w-full flex-1 flex-col' : undefined}>
          {path !== '' && !editor && (
            <div className="sticky top-0 z-10 flex min-h-12 items-center justify-between gap-2 border-b border-border bg-page px-4 py-2">
              <p className="min-w-0 truncate font-mono text-xs text-muted-foreground">
                docs/{path}
              </p>
              {doc && board?.readOnly === false && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={startEditing}
                  onMouseEnter={warmEditor}
                  onFocus={warmEditor}
                  disabled={loadingSource}
                >
                  {loadingSource && <Loader2Icon className="animate-spin" />}
                  Edit
                </Button>
              )}
            </div>
          )}
          {doc && editor && (
            <Suspense fallback={<p className="text-sm text-muted-foreground">Loading editor…</p>}>
              <DocEditor path={path} source={editor} onClose={() => setEditing(null)} />
            </Suspense>
          )}
          {!editor && (
            <div className="mx-auto max-w-3xl p-6">
              {path === '' && paths && <p className="text-sm text-muted-foreground">Pick a doc</p>}
              {path !== '' && isPending && !error && (
                <Loader2Icon
                  aria-label="Loading doc"
                  className="mx-auto mt-16 animate-spin text-muted-foreground"
                />
              )}
              {error && (
                <p className="text-sm text-muted-foreground">
                  This doc does not exist in the repo.
                </p>
              )}
              {doc && (
                <>
                  <RichContent html={doc.html} taskIds={board?.cards.map((c) => c.id) ?? []} />
                  {doc.backlinks.length > 0 && (
                    <footer className="mt-8 flex flex-wrap items-center gap-2 border-t border-border pt-4 text-xs text-muted-foreground">
                      Mentioned by{' '}
                      {doc.backlinks.map((id) => (
                        <TaskLink key={id} id={id} />
                      ))}
                    </footer>
                  )}
                </>
              )}
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
