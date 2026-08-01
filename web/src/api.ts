const BASE = ''

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  overview: () => request<any>('/api/overview/status'),

  tasks: {
    list: (status?: string) =>
      request<any[]>(`/api/tasks${status && status !== 'all' ? `?status=${encodeURIComponent(status)}` : ''}`),
    get: (id: number) => request<any>(`/api/tasks/${id}`),
    create: (data: any) => request<any>('/api/tasks', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: number, data: any) => request<any>(`/api/tasks/${id}`, { method: 'PATCH', body: JSON.stringify(data) }),
    delete: (id: number) => request(`/api/tasks/${id}`, { method: 'DELETE' }),
    deleteCompleted: () => request('/api/tasks/completed', { method: 'DELETE' }),
    retry: (id: number) => request<any>(`/api/tasks/${id}/retry`, { method: 'POST' }),
    batch: (ids: number[], action: string) =>
      request('/api/tasks/batch', { method: 'POST', body: JSON.stringify({ ids, action }) }),
  },

  wizard: {
    parse: (path: string) => request<any>('/api/wizard/parse', { method: 'POST', body: JSON.stringify({ path }) }),
    fileStreams: (path: string) => request<any>(`/api/wizard/file/${encodeURIComponent(path)}/streams`),
  },

  configs: {
    list: () => request<any[]>('/api/configs'),
    get: (id: number) => request<any>(`/api/configs/${id}`),
    create: (data: any) => request<any>('/api/configs', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: number, data: any) => request<any>(`/api/configs/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id: number) => request(`/api/configs/${id}`, { method: 'DELETE' }),
  },

  presets: () => request<any[]>('/api/presets'),

  settings: {
    list: () => request<Record<string,string>>('/api/settings'),
    update: (data: Record<string,string>) =>
      request('/api/settings', { method: 'PATCH', body: JSON.stringify(data) }),
    gpuInfo: () => request<any[]>('/api/settings/gpu-info'),
  },

  preview: {
    create: (data: any) => request<any>('/api/preview', { method: 'POST', body: JSON.stringify(data) }),
    status: (id: number) => request<any>(`/api/preview/${id}/status`),
    delete: (id: number) => request(`/api/preview/${id}`, { method: 'DELETE' }),
    downloadUrl: (id: number, token: string) => `${BASE}/api/preview/${id}/download?token=${token}`,
  },

  logs: {
    get: (lines?: number, level?: string) => {
      const params = new URLSearchParams()
      if (lines) params.set('lines', String(lines))
      if (level && level !== 'all') params.set('level', level)
      return request<any>(`/api/logs?${params.toString()}`)
    },
    downloadUrl: () => `${BASE}/api/logs/download`,
  },
}
