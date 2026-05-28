import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchDelivery, retryDelivery } from '../api'
import { StatusBadge } from '../components/StatusBadge'

export function DeliveryDetailPage() {
  const { id } = useParams<{ id: string }>()
  const qc = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['delivery', id],
    queryFn: () => fetchDelivery(id!),
    enabled: !!id,
  })

  const retry = useMutation({
    mutationFn: () => retryDelivery(id!),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['delivery', id] }),
  })

  if (isLoading) return <div className="px-6 py-8 text-gray-500 text-sm">Loading…</div>
  if (!data) return <div className="px-6 py-8 text-red-400 text-sm">Delivery not found.</div>

  const { delivery, attempts } = data

  return (
    <div className="px-6 py-6 max-w-4xl">
      <div className="flex items-center gap-3 mb-6">
        <Link to="/deliveries" className="text-gray-500 hover:text-gray-200 text-sm">← Deliveries</Link>
        <span className="text-gray-700">/</span>
        <span className="font-mono text-sm text-gray-300">{delivery.id}</span>
      </div>

      <div className="flex items-start justify-between mb-6">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <StatusBadge status={delivery.status} />
            <span className="text-sm text-gray-500">{delivery.attempt_count} attempt{delivery.attempt_count !== 1 ? 's' : ''}</span>
          </div>
          <div className="flex gap-4 text-xs text-gray-500 font-mono mt-2">
            <span>event: <Link to={`/events/${delivery.event_id}`} className="text-gray-400 hover:text-emerald-400">{delivery.event_id.slice(0, 12)}…</Link></span>
            <span>endpoint: {delivery.endpoint_id.slice(0, 12)}…</span>
          </div>
        </div>
        {(delivery.status === 'failed' || delivery.status === 'retrying') && (
          <button
            onClick={() => retry.mutate()}
            disabled={retry.isPending}
            className="px-3 py-1.5 rounded bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium disabled:opacity-50 transition-colors"
          >
            {retry.isPending ? 'Retrying…' : 'Retry Now'}
          </button>
        )}
      </div>

      <h2 className="text-sm font-medium text-gray-300 mb-3">Attempts</h2>

      {attempts.length === 0 && (
        <p className="text-sm text-gray-500">No attempts yet.</p>
      )}

      <div className="space-y-3">
        {attempts.map(a => (
          <div key={a.id} className="rounded-lg border border-gray-800 overflow-hidden">
            <div className="px-4 py-2.5 bg-gray-900/50 border-b border-gray-800 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <span className="text-xs font-medium text-gray-400">Attempt #{a.attempt}</span>
                <StatusBadge status={a.status} />
                {a.response_status && (
                  <span className="text-xs font-mono text-gray-400">HTTP {a.response_status}</span>
                )}
              </div>
              <div className="flex items-center gap-3 text-xs text-gray-500">
                {a.latency_ms != null && <span>{a.latency_ms}ms</span>}
                <span>{new Date(a.attempted_at).toLocaleTimeString()}</span>
              </div>
            </div>

            <div className="grid grid-cols-2 divide-x divide-gray-800">
              <div className="px-4 py-3">
                <p className="text-xs text-gray-500 mb-1.5 uppercase tracking-wider">Request</p>
                <pre className="text-xs font-mono text-gray-400 overflow-auto max-h-32">
                  {a.request_body ?? '(empty)'}
                </pre>
              </div>
              <div className="px-4 py-3">
                <p className="text-xs text-gray-500 mb-1.5 uppercase tracking-wider">Response</p>
                <pre className="text-xs font-mono text-gray-400 overflow-auto max-h-32">
                  {a.response_body ?? '(empty)'}
                </pre>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
