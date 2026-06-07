import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { api, type Health, type DataCoverage, type PaperStatus, type Trade } from "@/lib/api"
import { useWS } from "@/hooks/useWebSocket"
import { Activity, Database, TrendingUp, BarChart3, Clock, AlertTriangle } from "lucide-react"

function StatCard({ title, value, icon: Icon, subtitle, badge }: {
  title: string; value: string; icon: React.FC<{ className?: string }>; subtitle?: string; badge?: { label: string; variant: "success" | "warning" | "destructive" | "default" }
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {subtitle && <p className="text-xs text-muted-foreground mt-1">{subtitle}</p>}
        {badge && <Badge variant={badge.variant} className="mt-2">{badge.label}</Badge>}
      </CardContent>
    </Card>
  )
}

export function Dashboard() {
  const [health, setHealth] = useState<Health | null>(null)
  const [overview, setOverview] = useState<DataCoverage[]>([])
  const [paper, setPaper] = useState<PaperStatus | null>(null)
  const [trades, setTrades] = useState<Trade[]>([])
  const [loading, setLoading] = useState(true)

  // Live updates from WebSocket
  const wsAccount = useWS("account")
  const wsHealth = useWS("health")

  useEffect(() => {
    Promise.all([
      api.health(),
      api.overview(),
      api.paperStatus(),
      api.paperTrades(5),
    ]).then(([h, o, p, t]) => {
      setHealth(h)
      setOverview(o)
      setPaper(p)
      setTrades(t)
    }).catch(console.error).finally(() => setLoading(false))
  }, [])

  // Merge live WS data
  const liveBalance = wsAccount?.balance ?? paper?.balance ?? 10000
  const liveEquity = wsAccount?.equity ?? paper?.equity ?? 10000
  const liveDayPnL = wsAccount?.day_pnl ?? paper?.day_pnl ?? 0
  const liveDayTrades = wsAccount?.day_trades ?? paper?.day_trades ?? 0
  const liveCollector = wsHealth?.collector ?? health?.collector ?? "unknown"

  if (loading) return (
    <div className="space-y-4">
      <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-32 rounded-xl" />)}
      </div>
    </div>
  )

  const candles = overview.find((o) => o.table === "candles_15m")
  const features = overview.find((o) => o.table === "feature_values")
  const labels = overview.find((o) => o.table === "training_labels")

  const healthOk = liveCollector === "ok" || liveCollector === "healthy"
  const healthStatus = healthOk ? "success" : "destructive"

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <div className="flex items-center gap-2">
          <div className={`h-2 w-2 rounded-full ${healthOk ? "bg-emerald-500" : "bg-red-500"}`} />
          <Badge variant={healthStatus}>
            {healthOk ? "Live" : "Disconnected"}
          </Badge>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Total Candles (15m)" value={candles?.count.toLocaleString() ?? "0"} icon={Database} subtitle={`${candles?.span_days ?? 0}d span`} />
        <StatCard title="Features" value={features?.count.toLocaleString() ?? "0"} icon={BarChart3} subtitle={`${features?.span_days ?? 0}d span`} />
        <StatCard title="Labels" value={labels?.count.toLocaleString() ?? "0"} icon={Activity} />
        <StatCard title="Balance" value={`$${liveBalance.toLocaleString()}`} icon={TrendingUp}
          badge={{ label: "Running", variant: "success" }}
        />
        <StatCard title="Day PnL" value={`$${liveDayPnL.toFixed(2)}`} icon={TrendingUp}
          badge={{ label: liveDayPnL >= 0 ? "Profitable" : "Loss", variant: liveDayPnL >= 0 ? "success" : "destructive" }}
        />
        <StatCard title="Uptime" value={`${health?.uptime_hours.toFixed(1) ?? 0}h`} icon={Clock} subtitle="Collector" />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-4 w-4" /> System Health
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm">Database</span>
              <Badge variant={health?.db === "ok" ? "success" : "destructive"}>{health?.db ?? "unknown"}</Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm">Collector</span>
              <Badge variant={health?.collector === "ok" ? "success" : "destructive"}>{health?.collector ?? "unknown"}</Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm">Paper Trader</span>
              <Badge variant={health?.paper_trader === "running" ? "success" : "warning"}>{health?.paper_trader ?? "unknown"}</Badge>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <TrendingUp className="h-4 w-4" /> Recent Trades
            </CardTitle>
          </CardHeader>
          <CardContent>
            {trades.length === 0 ? (
              <div className="flex flex-col items-center gap-2 py-6 text-muted-foreground">
                <AlertTriangle className="h-8 w-8" />
                <p className="text-sm">No trades yet</p>
              </div>
            ) : (
              <div className="space-y-2">
                {trades.map((t) => (
                  <div key={t.id} className="flex items-center justify-between rounded-lg border p-2">
                    <div>
                      <p className="text-sm font-medium">{t.symbol} {t.direction}</p>
                      <p className="text-xs text-muted-foreground">{new Date(t.exit_ts).toLocaleString()}</p>
                    </div>
                    <div className="text-right">
                      <p className={`text-sm font-bold ${t.net_pnl >= 0 ? "text-emerald-500" : "text-red-500"}`}>
                        ${t.net_pnl.toFixed(2)}
                      </p>
                      <p className="text-xs text-muted-foreground">{t.exit_reason}</p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
