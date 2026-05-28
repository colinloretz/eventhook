import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchEvent, fetchDeliveries, replayEvent } from '../api'
import { StatusBadge } from '../components/StatusBadge'

export function EventDetailPage() {
  const { id } = useParams<{ id: string }>()
  const qc = useQueryClient()

  const { data: event, isLoading } = useQuery({
    queryKey: ['event', id],
    queryFn: () => fetchEvent(id!),
    enabled: !!id,
  })

  const { data: deliveries = [] } = useQuery({
    queryKey: ['deliveries', { event_id: id }],
    queryFn: () => fetchDeliveries({ event_id: id! }),
    enabled: !!id,
  })

  const replay = useMutation({
    mutationFn: () => replayEvent(id!),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['events'] }),
  })

  if (isLoading) return <div className="px-6 py-8 text-gray-500 text-sm">Loading…</div>
  if (!event) return <div className="px-6 py-8 text-red-400 text-sm">Event not found.</div>

  return (
    <div className="px-6 py-6 max-w-4xl">
      <div className="flex items-center gap-3 mb-6">
        <Link to="/" className="text-gray-500 hover:text-gray-200 text-sm">← Events</Link>
        <span className="text-gray-700">/</span>
        <span className="font-mono text-sm text-gray-300">{event.id}</span>
      </div>

      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="text-lg font-semibold font-mono">{event.event_type}</h1>
          <p className="text-sm text-gray-500 mt-0.5">
            {new Date(event.received_at).toLocaleString()}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <StatusBadge status={event.status} />
          <button
            onClick={() => replay.mutate()}
            disabled={replay.isPending}
            className="px-3 py-1.5 rounded bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium disabled:opacity-50 transition-colors"
          >
            {replay.isPending ? 'Replaying…' : 'Replay'}
          </button>
        </div>
      </div>

      <Section title="Payload">
        <pre className="text-xs font-mono text-gray-300 overflow-auto">
          {JSON.stringify(event.payload, null, 2)}
        </pre>
      </Section>

      <Section title="Headers">
        <pre className="text-xs font-mono text-gray-300 overflow-auto">
          {JSON.stringify(event.headers, null, 2)}
        </pre>
      </Section>

      {deliveries.length > 0 && (
        <Section title="Deliveries">
          <div className="divide-y divide-gray-800">
            {deliveries.map(d => (
              <div key={d.id} className="py-2.5 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <StatusBadge status={d.status} />
                  <span className="font-mono text-xs text-gray-400">{d.endpoint_id}</span>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-xs text-gray-500">{d.attempt_count} attempt{d.attempt_count !== 1 ? 's' : ''}</span>
                  <Link to={`/deliveries/${d.id}`} className="text-xs text-emerald-400 hover:underline">
                    View →
                  </Link>
                </div>
              </div>
            ))}
          </div>
        </Section>
      )}
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="mb-5 rounded-lg border border-gray-800 overflow-hidden">
      <div className="px-4 py-2 bg-gray-900/50 border-b border-gray-800">
        <h2 className="text-xs font-medium text-gray-400 uppercase tracking-wider">{title}</h2>
      </div>
      <div className="px-4 py-3 bg-gray-900/20 max-h-64 overflow-auto">
        {children}
      </div>
    </div>
  )
}
