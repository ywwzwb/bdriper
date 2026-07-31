type EventHandler = (data: any) => void

const handlers = new Map<string, Set<EventHandler>>()

let ws: WebSocket | null = null

export function connect() {
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = new WebSocket(`${protocol}//${location.host}/ws/events`)

  ws.onmessage = (e) => {
    try {
      const event = JSON.parse(e.data)
      const hs = handlers.get(event.type)
      if (hs) hs.forEach(h => h(event.data))
    } catch { /* ignore malformed messages */ }
  }

  ws.onclose = () => setTimeout(connect, 3000)
}

export function on(event: string, handler: EventHandler) {
  if (!handlers.has(event)) handlers.set(event, new Set())
  handlers.get(event)!.add(handler)
  return () => handlers.get(event)?.delete(handler)
}

export function connectLogs(handler: (line: string) => void): () => void {
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const lws = new WebSocket(`${protocol}//${location.host}/ws/logs`)

  lws.onmessage = (e) => {
    try {
      handler(e.data)
    } catch { /* ignore */ }
  }

  lws.onclose = () => setTimeout(() => connectLogs(handler), 3000)

  return () => lws.close()
}
