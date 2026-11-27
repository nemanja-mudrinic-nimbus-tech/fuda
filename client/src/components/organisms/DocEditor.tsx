import MDEditor from '@uiw/react-md-editor/nohighlight'
import '@uiw/react-md-editor/markdown-editor.css'
import { useBlocker } from '@tanstack/react-router'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { useLeaveWarning } from '@/hooks/useLeaveWarning'
import { useTheme } from '@/hooks/useTheme'
import { ApiError, type DocSource } from '@/lib/api'
import { useSaveDoc } from '@/lib/queries'

const leavePrompt = 'You have unsaved changes to this doc. Leave without saving?'

function failureText(error: unknown): string {
  if (error instanceof ApiError && error.status === 409) return error.message
  if (error instanceof ApiError && error.status === 401) {
    return 'Your login expired. Copy your changes, log in again, then reload the page.'
  }
  return error instanceof ApiError ? error.message : 'The doc was not saved.'
}

export default function DocEditor({
  path,
  source,
  onClose,
}: {
  path: string
  source: DocSource
  onClose: () => void
}) {
  const [draft, setDraft] = useState(source.content)
  const save = useSaveDoc()
  const { resolved } = useTheme()
  const dirty = draft !== source.content
  useLeaveWarning(dirty)
  useBlocker({
    shouldBlockFn: () => !window.confirm(leavePrompt),
    disabled: !dirty,
    enableBeforeUnload: false,
  })

  const submit = () =>
    save.mutate({ path, version: source.version, content: draft }, { onSuccess: onClose })

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      <div className="flex items-center justify-between gap-2 px-6 pt-4">
        <p className="min-w-0 truncate font-mono text-xs text-muted-foreground">
          Editing: docs/{path}
        </p>
        <div className="flex shrink-0 gap-2">
          <Button variant="outline" size="sm" onClick={onClose} disabled={save.isPending}>
            Cancel
          </Button>
          <Button size="sm" onClick={submit} disabled={!dirty || save.isPending}>
            {save.isPending ? 'Saving…' : 'Save'}
          </Button>
        </div>
      </div>
      {save.error && (
        <p role="alert" className="px-6 text-sm text-destructive">
          {failureText(save.error)}
        </p>
      )}
      <div
        data-color-mode={resolved}
        className="relative min-h-0 flex-1 border-t border-border [&_.w-md-editor]:rounded-none! [&_.w-md-editor]:shadow-none!"
      >
        <div className="absolute inset-0">
          <MDEditor
            value={draft}
            onChange={(value) => setDraft(value ?? '')}
            height="100%"
            visibleDragbar={false}
          />
        </div>
      </div>
    </div>
  )
}
