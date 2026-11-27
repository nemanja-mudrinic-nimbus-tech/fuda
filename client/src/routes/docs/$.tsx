import { createFileRoute } from '@tanstack/react-router'
import { lazy, Suspense, useState } from 'react'
import { toast } from 'sonner'
import { RichContent } from '@/components/molecules/RichContent'
import { TaskLink } from '@/components/atoms/TaskLink'
import { DocTree } from '@/components/organisms/DocTree'
import { Button } from '@/components/ui/button'
import { ApiError, api, type DocSource } from '@/lib/api'
import { useBoard, useDoc, useDocList } from '@/lib/queries'

const DocEditor = lazy(() => import('@/components/organisms/DocEditor'))

export const Route = createFileRoute('/docs/$')({
  component: DocPage,
})

const defaultDoc = 'board/README.md'

function DocPage() {
  const { _splat: requested = '' } = Route.useParams()
  const { data: paths } = useDocList()
  const { data: board } = useBoard()
  const path = requested || (paths?.includes(defaultDoc) ? defaultDoc : '')
  const { data: doc, error } = useDoc(path)
  const [editing, setEditing] = useState<{ path: string; source: DocSource } | null>(null)
  const [loadingSource, setLoadingSource] = useState(false)
  if (editing && editing.path !== path) setEditing(null)
  const editor = editing?.path === path ? editing.source : undefined

  const startEditing = async () => {
    setLoadingSource(true)
    try {
      setEditing({ path, source: await api.docSource(path) })
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : 'The doc could not be opened for editing')
    } finally {
      setLoadingSource(false)
    }
  }

  return (
    <div className="flex min-h-0 flex-1">
      <DocTree paths={paths ?? []} openPath={path} />
      <main className="flex-1 overflow-y-auto">
        <div className="mx-auto max-w-3xl p-6">
          {path === '' && paths && <p className="text-sm text-muted-foreground">Pick a doc</p>}
          {path !== '' && (
            <div className="mb-4 flex items-center justify-between gap-2">
              <p className="font-mono text-xs text-muted-foreground">docs/{path}</p>
              {doc && !editor && board?.readOnly === false && (
                <Button variant="outline" size="sm" onClick={startEditing} disabled={loadingSource}>
                  Edit
                </Button>
              )}
            </div>
          )}
          {error && (
            <p className="text-sm text-muted-foreground">This doc does not exist in the repo.</p>
          )}
          {doc && editor && (
            <Suspense fallback={<p className="text-sm text-muted-foreground">Loading editor…</p>}>
              <DocEditor path={path} source={editor} onClose={() => setEditing(null)} />
            </Suspense>
          )}
          {doc && !editor && (
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
      </main>
    </div>
  )
}
