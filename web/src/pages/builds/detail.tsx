import { useParams } from 'react-router'
import { BuildLogViewer } from '@/components/build-log-viewer'

export function BuildDetailPage() {
  const { id, buildId } = useParams()
  const projectId = Number(id)
  const buildIdNum = Number(buildId)

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">
          Build #{buildId}
        </h1>
        <p className="mt-1 text-muted-foreground">
          Build logs and status for project {id}.
        </p>
      </div>
      <BuildLogViewer
        buildId={buildIdNum}
        projectId={projectId}
        status="running"
      />
    </div>
  )
}
