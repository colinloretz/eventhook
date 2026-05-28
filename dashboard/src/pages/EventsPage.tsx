import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { fetchEvents } from '../api'
import { StatusBadge } from '../components/StatusBadge'
import type { Event } from '../types'

const STATUS_OPTIONS = ['', 'pending', 'delivered', 'failed', 'ignored']

function formatTime(iso: string) {
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function payloadPreview(payload: Record<string, unknown>) {
  return JSON.stringify(payload).slice(0, 80)
}

export function EventsPage() {
  const [status, setStatus] = useState('')
  const [eventType, setEventType] = useState('')

  const params: Record<string, string> = {}
  if (status) params.status = status
  if (eventType) params.event_type = eventType

  const { data: events = [], isLoading } = useQuery({
    queryKey: ['events', params],
    queryFn: () => fetchEvents(params),
    refetchInterval: 2000,
  })

  return (
    <div className="px-6 py-6">
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-lg font-semibold">Event Stream</h1>
        <div className="flex gap-3">
          <input
            type="text"
            placeholder="Filter by event type…"
            value={eventType}
            onChange={e => setEventType(e.target.value)}
            className="bg-gray-900 border border-gray-700 rounded px-3 py-1.5 text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 w-52"
          />
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
      </div>

      <div className="rounded-lg border border-gray-800 overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-gray-800 bg-gray-900/50">
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium w-24">Time</th>
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium w-32">Status</th>
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium w-52">Event Type</th>
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium">Payload</th>
            </tr>
          </thead>
          <tbody>
            {isLoading && (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-gray-500 text-sm">
                  Loading…
                </td>
              </tr>
            )}
            {!isLoading && events.length === 0 && (
              <tr>
                <td colSpan={4} className="px-4 py-12 text-center text-gray-500 text-sm">
                  No events yet. Waiting for events…
                </td>
              </tr>
            )}
            {events.map((ev: Event) => (
              <tr
                key={ev.id}
                className="border-b border-gray-800/50 hover:bg-gray-800/30 transition-colors cursor-pointer"
              >
                <td className="px-4 py-2.5 font-mono text-xs text-gray-500">
                  {formatTime(ev.received_at)}
                </td>
                <td className="px-4 py-2.5">
                  <StatusBadge status={ev.status} />
                </td>
                <td className="px-4 py-2.5">
                  <Link to={`/events/${ev.id}`} className="text-gray-200 hover:text-emerald-400 font-mono text-xs">
                    {ev.event_type}
                  </Link>
                </td>
                <td className="px-4 py-2.5 font-mono text-xs text-gray-500 truncate max-w-xs">
                  {payloadPreview(ev.payload)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="mt-3 flex items-center gap-2">
        <span className="inline-block w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
        <span className="text-xs text-gray-500">Live — polling every 2s</span>
      </div>
    </div>
  )
}
