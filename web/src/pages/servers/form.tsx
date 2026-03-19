import { useParams } from 'react-router'

export function ServerFormPage() {
  const { id } = useParams()
  const isEdit = !!id

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">
          {isEdit ? 'Edit Server' : 'New Server'}
        </h1>
        <p className="mt-1 text-muted-foreground">
          {isEdit
            ? 'Update your server configuration.'
            : 'Add a new deployment server.'}
        </p>
      </div>
      <div className="rounded-lg border border-zinc-200 bg-white p-8 dark:border-zinc-800 dark:bg-zinc-900">
        <p className="text-center text-sm text-muted-foreground">
          Server form coming soon
        </p>
      </div>
    </div>
  )
}
