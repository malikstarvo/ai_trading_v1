import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { Button } from "@/components/ui/button"
import { api, type Health, type DataOverview } from "@/lib/api"
import { useWS } from "@/hooks/useWebSocket"
import {
  Activity, Database, AlertTriangle, CircleCheck, CircleAlert,
  TrendingUp, Clock, Zap, BarChart3, ArrowRight, RefreshCw
} from "lucide-react"

function PipelineStage({
  number,
  label,
  status,
  description,
  isLast
}: {
  number: number
  label: string
  status: "complete" | "active" | "pending"
  description: string
  isLast?: boolean
}) {
  const statusColors = {
    complete: "bg-emerald-500 text-white",
    active: "bg-primary text-primary-foreground",
    pending: "bg-muted text-muted-foreground"
  }

  return (
    <div className="flex items-center gap-3">
      <div className="flex flex-col items-center">
        <div className={cn(
          "flex h-10 w-10 items-center justify-center rounded-full text-sm font-bold",
          statusColors[status]
        )}>
          {status === "complete" ? <CircleCheck className="h-5 w-5" /> : number}
        </div>
        {!isLast && (
          <div className={cn(
            "h-8 w-0.5 mt-1",
            status === "complete" ? "bg-emerald-500" : "bg-muted"
          )} />
        )}
      </div>
      <div className="pb-4">
        <p className="font-medium text-sm">{label}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
    </div>
  )
}

function DataReadinessCard({ symbol, tf, days, target, oiNan, lsNan }: {
  symbol: string
  tf: string
  days: number
  target: number
  oiNan: string | null
  lsNan: string | null
}) {
  const pct = Math.min((days / target) * 100, 100)
  const isReady = pct >= 90 && parseFloat(oiNan || "100") < 30 && parseFloat(lsNan || "100") < 30
  const color = isReady ? "bg-emerald-500" : pct >= 50 ? "bg-amber-500" : "bg-red-500"

  return (
    <Card className="relative overflow-hidden">
      <div className={cn("absolute top-0 left-0 h-1 w-full", color)} />
      <CardContent className="pt-4 pb-3">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-sm">{symbol}</span>
            <Badge variant="secondary" className="text-xs">{tf}</Badge>
          </div>
          {isReady ? (
            <CircleCheck className="h-4 w-4 text-emerald-500" />
          ) : (
            <CircleAlert className="h-4 w-4 text-amber-500" />
          )}
        </div>
        <div className="space-y-2">
          <div>
            <div className="flex justify-between text-xs mb-1">
              <span className="text-muted-foreground">Progress</span>
              <span className="font-medium">{days}/{target}d</span>
            </div>
            <div className="h-2 w-full rounded-full bg-muted">
              <div className={cn("h-full rounded-full transition-all", color)} style={{ width: `${pct}%` }} />
            </div>
          </div>
          <div className="flex gap-2 text-[11px] text-muted-foreground">
            <span>OI NaN: {oiNan || "—"}%</span>
            <span>LS NaN: {lsNan || "—"}%</span>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

function AlertBanner({ type, title, message }: {
  type: "error" | "warning" | "info"
  title: string
  message: string
}) {
  const colors = {
    error: "bg-red-500/10 text-red-700 border-red-200",
    warning: "bg-amber-500/10 text-amber-700 border-amber-200",
    info: "bg-blue-500/10 text-blue-700 border-blue-200"
  }

  return (
    <div className={cn("flex items-start gap-3 rounded-lg border p-3", colors[type])}>
      <AlertTriangle className="h-5 w-5 shrink-0 mt-0.5" />
      <div>
        <p className="font-medium text-sm">{title}</p>
        <p className="text-xs opacity-80">{message}</p>
      </div>
    </div>
  )
}

function cn(...inputs: (string | undefined | false | null)[]) {
  return inputs.filter(Boolean).join(" ")
}

export function Overview() {
  const [health, setHealth] = useState<Health | null>(null)
  const [data, setData] = useState<DataOverview | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const wsHealth = useWS("health")
  const wsAccount = useWS("account")

  const fetchData = async () => {
    try {
      setError(false)
      const [h, d] = await Promise.all([api.health(), api.dataOverview()])
      setHealth(h)
      setData(d)
    } catch {
      setError(true)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 30000)
    return () => clearInterval(interval)
  }, [])

  const liveCollector = wsHealth?.collector ?? health?.collector ?? "unknown"
  const healthOk = liveCollector === "ok" || liveCollector === "healthy"

  // Calculate alerts
  const alerts: { type: "error" | "warning" | "info"; title: string; message: string }[] = []

  if (!healthOk) {
    alerts.push({ type: "error", title: "Collector Disconnected", message: "WebSocket connection to Bybit is down. Check logs." })
  }

  if (data) {
    const anyStale = data.symbols.some(s => s.candle_progress_pct < 50)
    if (anyStale) {
      alerts.push({ type: "warning", title: "Data Collection Incomplete", message: "Some symbols are below 50% of the 60-day target." })
    }

    const highNaN = data.symbols.some(s => s.oi_nan_pct && parseFloat(s.oi_nan_pct) > 90)
    if (highNaN) {
      alerts.push({ type: "warning", title: "High OI NaN Rate", message: "Some 1h timeframes have >90% OI NaN. Wait for WS accumulation." })
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <div className="grid gap-4 md:grid-cols-3">
          <Skeleton className="h-32" />
          <Skeleton className="h-32" />
          <Skeleton className="h-32" />
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold tracking-tight">Overview</h1>
          <Badge variant="destructive">Disconnected</Badge>
        </div>
        <AlertBanner type="error" title="Connection Error" message="Failed to fetch data from API. Please check if the API server is running." />
        <Button onClick={fetchData}><RefreshCw className="mr-2 h-4 w-4" /> Retry</Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Overview</h1>
          <p className="text-sm text-muted-foreground mt-1">Mission Control — AI Trading Pipeline</p>
        </div>
        <div className="flex items-center gap-2">
          <div className={cn("h-2.5 w-2.5 rounded-full", healthOk ? "bg-emerald-500 animate-pulse" : "bg-red-500")} />
          <Badge variant={healthOk ? "success" : "destructive"}>
            {healthOk ? "Live" : "Disconnected"}
          </Badge>
        </div>
      </div>

      {/* Alerts */}
      {alerts.length > 0 && (
        <div className="space-y-2">
          {alerts.map((alert, i) => (
            <AlertBanner key={i} {...alert} />
          ))}
        </div>
      )}

      {/* Pipeline Status */}
      <div className="grid gap-6 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Zap className="h-4 w-4" /> Pipeline Status
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <PipelineStage
                number={1}
                label="Data Collection"
                status="active"
                description={`${data?.symbols[0]?.candle_days_span || 0}/60 days collected`}
              />
              <PipelineStage
                number={2}
                label="Model Training"
                status="pending"
                description="Waiting for 60 days of data"
              />
              <PipelineStage
                number={3}
                label="Live Trading"
                status="pending"
                description="Requires trained model"
                isLast
              />
            </div>
          </CardContent>
        </Card>

        {/* Quick Stats */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <BarChart3 className="h-4 w-4" /> Quick Stats
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Total Candles</span>
              <span className="font-mono font-medium">{data?.symbols.reduce((sum, s) => sum + s.candle_rows, 0).toLocaleString() || 0}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Latest Candle</span>
              <span className="font-mono text-xs">{data?.symbols[0]?.candle_last_ts?.slice(0, 16) || "—"}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Feature Lag</span>
              <span className="font-mono text-xs">{data?.symbols[0]?.feature_last_ts?.slice(0, 16) || "—"}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Target</span>
              <span className="font-mono font-medium">{data?.target_days || 60} days</span>
            </div>
          </CardContent>
        </Card>

        {/* System Health Mini */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Activity className="h-4 w-4" /> System Health
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm">Database</span>
              <div className="flex items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-emerald-500" />
                <span className="text-xs font-medium">OK</span>
              </div>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm">Collector</span>
              <div className="flex items-center gap-2">
                <div className={cn("h-2 w-2 rounded-full", healthOk ? "bg-emerald-500" : "bg-red-500")} />
                <span className="text-xs font-medium">{healthOk ? "Connected" : "Down"}</span>
              </div>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm">Paper Trader</span>
              <div className="flex items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-emerald-500" />
                <span className="text-xs font-medium">Running</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Data Readiness Grid */}
      <div>
        <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
          <Database className="h-5 w-5" /> Data Readiness
        </h2>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {data?.symbols.map((s) => (
            <DataReadinessCard
              key={`${s.symbol}-${s.timeframe}`}
              symbol={s.symbol}
              tf={s.timeframe}
              days={s.candle_days_span}
              target={s.target_days}
              oiNan={s.oi_nan_pct}
              lsNan={s.ls_nan_pct}
            />
          ))}
        </div>
      </div>

      {/* Live Account (from WS) */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <TrendingUp className="h-4 w-4" /> Live Paper Trading
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-4">
            <div>
              <p className="text-sm text-muted-foreground">Balance</p>
              <p className="text-2xl font-bold">${(wsAccount?.balance ?? 10000).toLocaleString()}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Equity</p>
              <p className="text-2xl font-bold">${(wsAccount?.equity ?? 10000).toLocaleString()}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Day PnL</p>
              <p className={cn("text-2xl font-bold", (wsAccount?.day_pnl ?? 0) >= 0 ? "text-emerald-500" : "text-red-500")}>
                ${(wsAccount?.day_pnl ?? 0).toFixed(2)}
              </p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Day Trades</p>
              <p className="text-2xl font-bold">{wsAccount?.day_trades ?? 0}</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}


