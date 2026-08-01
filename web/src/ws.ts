type EventHandler = (data: any) => void
const handlers = new Map<string, Set<EventHandler>>()
let ws: WebSocket | null = null

export function connect() {
  if (ws && ws.readyState === WebSocket.OPEN) return
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = new WebSocket(`${protocol}//${location.host}/ws/events`)
  ws.onmessage = (e) => {
    try {
      const event = JSON.parse(e.data)
      const hs = handlers.get(event.type)
      if (hs) hs.forEach(h => h(event.data))
    } catch {}
  }
  ws.onclose = () => setTimeout(connect, 3000)
  ws.onerror = () => {}
}

export function on(event: string, handler: EventHandler) {
  if (!handlers.has(event)) handlers.set(event, new Set())
  handlers.get(event)!.add(handler)
  return () => { handlers.get(event)?.delete(handler) }
}
