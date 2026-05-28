import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchEndpoints, createEndpoint, updateEndpoint, deleteEndpoint } from '../api'
import type { Endpoint } from '../types'

type FormState = {
  url: string
  description: string
  secret: string
  enabled: boolean
  event_types: string
}

const empty: FormState = { url: '', description: '', secret: '', enabled: true, event_types: '' }

export function EndpointsPage() {
  const qc = useQueryClient()
  const [editing, setEditing] = useState<Endpoint | null>(null)
  const [creating, setCreating] = useState(false)
  const [form, setForm] = useState<FormState>(empty)

  const { data: endpoints = [], isLoading } = useQuery({
    queryKey: ['endpoints'],
    queryFn: fetchEndpoints,
  })

  const create = useMutation({
    mutationFn: () => createEndpoint({
      url: form.url,
      description: form.description || undefined,
      secret: form.secret,
      enabled: form.enabled,
      event_types: form.event_types ? form.event_types.split(',').map(s => s.trim()) : [],
    }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['endpoints'] }); closeForm() },
  })

  const update = useMutation({
    mutationFn: () => updateEndpoint(editing!.id, {
      url: form.url,
      description: form.description || undefined,
      secret: form.secret,
      enabled: form.enabled,
      event_types: form.event_types ? form.event_types.split(',').map(s => s.trim()) : [],
    }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['endpoints'] }); closeForm() },
  })

  const remove = useMutation({
    mutationFn: (id: string) => deleteEndpoint(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['endpoints'] }),
  })

  const openCreate = () => { setEditing(null); setForm(empty); setCreating(true) }
  const openEdit = (ep: Endpoint) => {
    setEditing(ep)
    setForm({
      url: ep.url,
      description: ep.description ?? '',
      secret: '',
      enabled: ep.enabled,
      event_types: (ep.event_types ?? []).join(', '),
    })
    setCreating(true)
  }
  const closeForm = () => { setCreating(false); setEditing(null); setForm(empty) }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    editing ? update.mutate() : create.mutate()
  }

  return (
    <div className="px-6 py-6">
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-lg font-semibold">Endpoints</h1>
        <button
          onClick={openCreate}
          className="px-3 py-1.5 rounded bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium transition-colors"
        >
          + New Endpoint
        </button>
      </div>

      {creating && (
        <form onSubmit={handleSubmit} className="mb-6 rounded-lg border border-gray-700 bg-gray-900/50 p-5 space-y-4">
          <h2 className="text-sm font-medium text-gray-200">{editing ? 'Edit Endpoint' : 'New Endpoint'}</h2>
          <div className="grid grid-cols-2 gap-4">
            <Field label="URL" required>
              <input
                type="url"
                required
                value={form.url}
                onChange={e => setForm(f => ({ ...f, url: e.target.value }))}
                className={inputCls}
                placeholder="https://your-app.com/webhooks"
              />
            </Field>
            <Field label="Secret" required>
              <input
                type="text"
                required
                value={form.secret}
                onChange={e => setForm(f => ({ ...f, secret: e.target.value }))}
                className={inputCls}
                placeholder="whsec_…"
              />
            </Field>
            <Field label="Description">
              <input
                type="text"
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                className={inputCls}
                placeholder="Optional"
              />
            </Field>
            <Field label="Event Types (comma-separated)">
              <input
                type="text"
                value={form.event_types}
                onChange={e => setForm(f => ({ ...f, event_types: e.target.value }))}
                className={inputCls}
                placeholder="payment.completed, user.created (empty = all)"
              />
            </Field>
          </div>
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={form.enabled}
              onChange={e => setForm(f => ({ ...f, enabled: e.target.checked }))}
              className="rounded border-gray-600 bg-gray-800 text-emerald-500 focus:ring-emerald-500"
            />
            <span className="text-sm text-gray-300">Enabled</span>
          </label>
          <div className="flex gap-2 pt-1">
            <button
              type="submit"
              disabled={create.isPending || update.isPending}
              className="px-4 py-1.5 rounded bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium disabled:opacity-50 transition-colors"
            >
              {create.isPending || update.isPending ? 'Saving…' : editing ? 'Save Changes' : 'Create'}
            </button>
            <button type="button" onClick={closeForm} className="px-4 py-1.5 rounded bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm transition-colors">
              Cancel
            </button>
          </div>
        </form>
      )}

      <div className="rounded-lg border border-gray-800 overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-gray-800 bg-gray-900/50">
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium">URL</th>
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium w-20">Status</th>
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium">Event Types</th>
              <th className="text-left px-4 py-2.5 text-xs text-gray-500 font-medium w-28"></th>
            </tr>
          </thead>
          <tbody>
            {isLoading && (
              <tr><td colSpan={4} className="px-4 py-8 text-center text-gray-500 text-sm">Loading…</td></tr>
            )}
            {!isLoading && endpoints.length === 0 && (
              <tr><td colSpan={4} className="px-4 py-12 text-center text-gray-500 text-sm">No endpoints yet. Create one to start receiving deliveries.</td></tr>
            )}
            {endpoints.map(ep => (
              <tr key={ep.id} className="border-b border-gray-800/50 hover:bg-gray-800/20 transition-colors">
                <td className="px-4 py-3">
                  <div className="font-mono text-xs text-gray-300">{ep.url}</div>
                  {ep.description && <div className="text-xs text-gray-500 mt-0.5">{ep.description}</div>}
                </td>
                <td className="px-4 py-3">
                  <span className={`inline-flex items-center rounded px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${ep.enabled ? 'bg-emerald-500/20 text-emerald-400 ring-emerald-500/30' : 'bg-gray-500/20 text-gray-400 ring-gray-500/30'}`}>
                    {ep.enabled ? 'enabled' : 'disabled'}
                  </span>
                </td>
                <td className="px-4 py-3 text-xs text-gray-400 font-mono">
                  {ep.event_types?.length ? ep.event_types.join(', ') : <span className="text-gray-600">all events</span>}
                </td>
                <td className="px-4 py-3">
                  <div className="flex gap-2 justify-end">
                    <button
                      onClick={() => openEdit(ep)}
                      className="text-xs text-gray-400 hover:text-gray-200 transition-colors"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => { if (confirm('Delete this endpoint?')) remove.mutate(ep.id) }}
                      className="text-xs text-red-500 hover:text-red-400 transition-colors"
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

const inputCls = 'w-full bg-gray-900 border border-gray-700 rounded px-3 py-1.5 text-sm text-gray-200 placeholder-gray-600 focus:outline-none focus:ring-1 focus:ring-emerald-500'

function Field({ label, required, children }: { label: string; required?: boolean; children: React.ReactNode }) {
  return (
    <div>
      <label className="block text-xs text-gray-400 mb-1.5">
        {label}{required && <span className="text-red-400 ml-0.5">*</span>}
      </label>
      {children}
    </div>
  )
}
