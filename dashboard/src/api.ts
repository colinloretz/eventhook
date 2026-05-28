import type { Delivery, DeliveryAttempt, Endpoint, Event, Source } from './types'

const API_KEY = (window as unknown as Record<string, string>).__EVENTHOOK_API_KEY__ ?? 'dev-api-key'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`/api/v1${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${API_KEY}`,
      ...options?.headers,
    },
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(`${res.status}: ${text}`)
  }
  return res.json() as Promise<T>
}

// Events
export const fetchEvents = (params?: Record<string, string>) => {
  const qs = params ? '?' + new URLSearchParams(params).toString() : ''
  return request<{ data: Event[] }>(`/events${qs}`).then(r => r.data)
}

export const fetchEvent = (id: string) =>
  request<Event>(`/events/${id}`)

export const replayEvent = (id: string) =>
  request<Event>(`/events/${id}/replay`, { method: 'POST' })

// Deliveries
export const fetchDeliveries = (params?: Record<string, string>) => {
  const qs = params ? '?' + new URLSearchParams(params).toString() : ''
  return request<{ data: Delivery[] }>(`/deliveries${qs}`).then(r => r.data)
}

export const fetchDelivery = (id: string) =>
  request<{ delivery: Delivery; attempts: DeliveryAttempt[] }>(`/deliveries/${id}`)

export const retryDelivery = (id: string) =>
  request<Delivery>(`/deliveries/${id}/retry`, { method: 'POST' })

// Endpoints
export const fetchEndpoints = () =>
  request<{ data: Endpoint[] }>('/endpoints').then(r => r.data)

export const fetchEndpoint = (id: string) =>
  request<Endpoint>(`/endpoints/${id}`)

export const createEndpoint = (data: Partial<Endpoint> & { secret: string }) =>
  request<Endpoint>('/endpoints', { method: 'POST', body: JSON.stringify(data) })

export const updateEndpoint = (id: string, data: Partial<Endpoint> & { secret: string }) =>
  request<Endpoint>(`/endpoints/${id}`, { method: 'PUT', body: JSON.stringify(data) })

export const deleteEndpoint = (id: string) =>
  request<void>(`/endpoints/${id}`, { method: 'DELETE' })

// Sources
export const fetchSources = () =>
  request<{ data: Source[] }>('/sources').then(r => r.data)
