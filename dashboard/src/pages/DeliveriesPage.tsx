import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { fetchDeliveries } from '../api'
import { StatusBadge } from '../components/StatusBadge'

const STATUS_OPTIONS = ['', 'pending', 'delivered', 'failed', 'retrying']

export function DeliveriesPage() {
  const [status, setStatus] = useState('')

  const params: Record<string, string> = {}
  if (status) params.status = status

  const { data: deliveries = [], isLoading } = useQuery({
    queryKey: ['deliveries', params],
    queryFn: () => fetchDeliveries(params),
    refetchInterval: 3000,
  })

  return (
    <div className="px-6 py-6">
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-lg font-semibold">Deliveries</h1>
        <select
          value={status}
          onChange={e => setStatus(e.target.value)}
          className="bg-gray-900 border border-gray-700 rounded px-3 py-1.5 text-sm text-gray-200 focus:outline-none focus:ring-1 focus:ring-emerald-500"
        >
          {STATUS_OPTIONS.map(s => (
            <option key={s} value={s}>{s || 'All statuses'}</option>
          ))}
        </select>
      </div>

      <div className="rounded-lg border border-gray-800 overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-gray-800 bg-gray-900/50">
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium w-24">Time</th>
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium w-32">Status</th>
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium">Event</th>
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium w-20">Attempts</th>
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium w-20">Response</th>
            </tr>
          </thead>
          <tbody>
            {isLoading && (
              <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-500 text-sm">Loading…</td></tr>
            )}
            {!isLoading && deliveries.length === 0 && (
              <tr><td colSpan={5} className="px-4 py-12 text-center text-gray-500 text-sm">No deliveries.</td></tr>
            )}
            {deliveries.map(d => (
              <tr key={d.id} className="border-b border-gray-800/50 hover:bg-gray-800/30 transition-colors">
                <td className="px-4 py-2.5 font-mono text-xs text-gray-500">
                  {new Date(d.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
                </td>
                <td className="px-4 py-2.5"><StatusBadge status={d.status} /></td>
                <td className="px-4 py-2.5">
                  <div className="flex gap-2 font-mono text-xs">
                    <Link to={`/deliveries/${d.id}`} className="text-gray-300 hover:text-emerald-400">
                      {d.id.slice(0, 8)}…
                    </Link>
                    <span className="text-gray-600">→</span>
                    <Link to={`/events/${d.event_id}`} className="text-gray-500 hover:text-gray-300">
                      event:{d.event_id.slice(0, 8)}…
                    </Link>
                  </div>
                </td>
                <td className="px-4 py-2.5 text-xs text-gray-400">{d.attempt_count}</td>
                <td className="px-4 py-2.5 text-xs text-gray-400">
                  {d.last_response_status ?? '—'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
