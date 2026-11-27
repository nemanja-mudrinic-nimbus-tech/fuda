import MDEditor from '@uiw/react-md-editor/nohighlight'
import '@uiw/react-md-editor/markdown-editor.css'
import { useBlocker } from '@tanstack/react-router'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
  const [message, setMessage] = useState(`docs: edit ${path}`)
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
    save.mutate({ path, version: source.version, content: draft, message }, { onSuccess: onClose })

  return (
    <div className="space-y-3">
      <div data-color-mode={resolved}>
        <MDEditor
          value={draft}
          onChange={(value) => setDraft(value ?? '')}
          height={520}
          visibleDragbar={false}
        />
      </div>
      {save.error && (
        <p role="alert" className="text-sm text-destructive">
          {failureText(save.error)}
        </p>
      )}
      <div className="flex items-center gap-2">
        <Input
          aria-label="Commit message"
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          className="flex-1"
        />
        <Button variant="outline" onClick={onClose} disabled={save.isPending}>
          Cancel
        </Button>
        <Button onClick={submit} disabled={!dirty || save.isPending}>
          {save.isPending ? 'Saving…' : 'Save'}
        </Button>
      </div>
    </div>
  )
}
