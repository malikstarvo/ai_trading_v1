import { useEffect, useRef, useState, useCallback } from "react"

export interface WSData {
  t?: number
  candle?: { time: string; close: number; volume: number }
  feature?: Record<string, number | string>
  account?: { balance: number; equity: number; day_pnl: number; day_trades: number }
  health?: { collector: string }
  heartbeat?: boolean
  pong?: boolean
}

type WSCallback = (data: WSData) => void

const listeners = new Set<WSCallback>()
let wsInstance: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectAttempt = 0

function getWSURL(): string {
  const proto = location.protocol === "https:" ? "wss:" : "ws:"
  return `${proto}//${location.host}/ws`
}

function connect() {
  if (wsInstance?.readyState === WebSocket.OPEN || wsInstance?.readyState === WebSocket.CONNECTING) {
    return
  }

  wsInstance = new WebSocket(getWSURL())

  wsInstance.onopen = () => {
    reconnectAttempt = 0
  }

  wsInstance.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data) as WSData
      listeners.forEach((cb) => cb(data))
    } catch {
      // ignore parse errors
    }
  }

  wsInstance.onclose = () => {
    wsInstance = null
    const delay = Math.min(1000 * 2 ** reconnectAttempt, 30000)
    reconnectAttempt++
    reconnectTimer = setTimeout(connect, delay)
  }

  wsInstance.onerror = () => {
    wsInstance?.close()
  }
}

function disconnect() {
  if (reconnectTimer) clearTimeout(reconnectTimer)
  reconnectTimer = null
  if (wsInstance) {
    wsInstance.onclose = null
    wsInstance.close()
    wsInstance = null
  }
  listeners.clear()
}

function subscribe(cb: WSCallback): () => void {
  listeners.add(cb)
  return () => { listeners.delete(cb) }
}

/**
 * Hook that provides a reactive snapshot of the latest WebSocket data.
 * Auto-connects on mount, auto-reconnects on disconnect.
 */
export function useWebSocket(): WSData | null {
  const [data, setData] = useState<WSData | null>(null)
  const cbRef = useRef<WSCallback>(setData)
  cbRef.current = setData

  useEffect(() => {
    connect()
    return subscribe((d) => cbRef.current(d))
  }, [])

  return data
}

/**
 * Hook with per-type selector — re-renders only when the selected type changes.
 */
export function useWS<T extends keyof WSData>(type: T): WSData[T] | null {
  const [value, setValue] = useState<WSData[T] | null>(null)

  useEffect(() => {
    connect()
    return subscribe((data) => {
      if (data[type] !== undefined) {
        setValue(data[type] as WSData[T])
      }
    })
  }, [type])

  return value
}
