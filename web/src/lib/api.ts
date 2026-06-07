const BASE = '/api'

async function fetchJSON<T>(url: string): Promise<T> {
  const res = await fetch(`${BASE}${url}`)
  if (!res.ok) throw new Error(`GET ${url} → ${res.status}`)
  return res.json()
}

// ── Types ──
export interface Health {
  db: string
  collector: string
  paper_trader: string
  uptime_hours: number
}

export interface DataCoverage {
  table: string
  count: number
  span_days: number | null
  first_ts: string | null
  last_ts: string | null
}

export interface Candle {
  time: string
  open: number
  high: number
  low: number
  close: number
  volume: number
}

export interface OrderflowPoint {
  time: string
  oi?: number
  oi_value_usd?: number
  funding_rate?: number
  buy_ratio?: number
  sell_ratio?: number
}

export interface Liquidation {
  time: string
  side: string
  size: number
  price: number
  value_usd: number
}

export interface FeatureNaN {
  column: string
  non_null: number
  total_rows: number
  nan_pct: number
}

export interface FeatureRanking {
  rank: number
  feature: string
  score: number
}

export interface FeatureRow {
  ts: string
  [key: string]: unknown
}

export interface LabelDist {
  horizon: number
  total: number
  success_rate: number
}

export interface PaperStatus {
  state: string
  symbol: string
  timeframe: string
  initial_capital: number
  balance: number
  equity: number
  day_pnl: number
  day_trades: number
  total_pnl: number
  uptime_hours: number
}

export interface Position {
  id: number
  symbol: string
  direction: string
  entry_price: number
  quantity: number
  stop_price: number
  open_ts: string
  bars_held: number
  status: string
  exit_price?: number
  net_pnl?: number
  return_pct?: number
  exit_reason?: string
}

export interface Trade {
  id: number
  symbol: string
  direction: string
  entry_ts: string
  exit_ts: string
  entry_price: number
  exit_price: number
  net_pnl: number
  return_pct: number
  holding_bars: number
  exit_reason: string
}

export interface AccountSnapshot {
  ts: string
  balance: number
  equity: number
  unrealized_pnl: number
  day_pnl: number
  day_trades: number
}

export interface ModelStatus {
  symbol: string
  timeframe: string
  models: {
    horizon: number
    exists: boolean
    auc: number | null
    trained_at: string | null
  }[]
}

export interface SystemInfo {
  collector: {
    running: boolean
    ws_connected: boolean
    uptime_hours: number
    last_heartbeat: string
  }
  db: {
    size_mb: number
    tables: { name: string; count: number }[]
  }
  crons: { schedule: string; command: string }[]
  paper_trader: {
    pid: number | null
    uptime_hours: number
  }
}

// ── API calls ──
export const api = {
  health: () => fetchJSON<Health>('/health'),
  overview: () => fetchJSON<DataCoverage[]>('/stats/overview'),
  candles: (symbol = 'BTCUSDT', tf = '15m', limit = 200) =>
    fetchJSON<Candle[]>(`/candles?symbol=${symbol}&tf=${tf}&limit=${limit}`),
  orderflow: (symbol = 'BTCUSDT', limit = 200) =>
    fetchJSON<OrderflowPoint[]>(`/orderflow?symbol=${symbol}&limit=${limit}`),
  liquidations: (symbol = 'BTCUSDT', limit = 100) =>
    fetchJSON<Liquidation[]>(`/liquidations?symbol=${symbol}&limit=${limit}`),
  featuresLatest: (symbol = 'BTCUSDT', tf = '15m') =>
    fetchJSON<FeatureRow>(`/features/latest?symbol=${symbol}&tf=${tf}`),
  featuresNaN: (symbol = 'BTCUSDT', tf = '15m') =>
    fetchJSON<FeatureNaN[]>(`/features/nan?symbol=${symbol}&tf=${tf}`),
  featuresRanking: () => fetchJSON<FeatureRanking[]>('/features/ranking'),
  labels: (symbol = 'BTCUSDT', tf = '15m') =>
    fetchJSON<LabelDist[]>(`/labels/distribution?symbol=${symbol}&tf=${tf}`),
  paperStatus: () => fetchJSON<PaperStatus>('/paper/status'),
  paperPositions: (status = '') =>
    fetchJSON<Position[]>(`/paper/positions?status=${status}`),
  paperTrades: (limit = 50) => fetchJSON<Trade[]>(`/paper/trades?limit=${limit}`),
  paperAccount: (limit = 200) => fetchJSON<AccountSnapshot[]>(`/paper/account?limit=${limit}`),
  modelStatus: () => fetchJSON<ModelStatus>('/model/status'),
  system: () => fetchJSON<SystemInfo>('/system'),
}
