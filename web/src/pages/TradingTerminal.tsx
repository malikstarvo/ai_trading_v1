import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { api, type PaperStatus, type Position, type Trade, type AccountSnapshot } from "@/lib/api"
import { useWS } from "@/hooks/useWebSocket"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { AreaChart, Area, BarChart, Bar, XAxis, YAxis, CartesianGrid } from "recharts"
import { DollarSign, TrendingUp, BarChart3, Clock, Target } from "lucide-react"

function cn(...inputs: (string | undefined | false | null)[]) {
  return inputs.filter(Boolean).join(" ")
}

export function TradingTerminal() {
  const [status, setStatus] = useState<PaperStatus | null>(null)
  const [positions, setPositions] = useState<Position[]>([])
  const [trades, setTrades] = useState<Trade[]>([])
  const [account, setAccount] = useState<AccountSnapshot[]>([])
  const [loading, setLoading] = useState(true)
  const wsAccount = useWS("account")

  useEffect(() => {
    Promise.all([
      api.paperStatus(),
      api.paperPositions(),
      api.paperTrades(100),
      api.paperAccount(200),
    ]).then(([s, p, t, a]) => {
      setStatus(s)
      setPositions(p)
      setTrades(t)
      setAccount(a)
    }).catch(console.error).finally(() => setLoading(false))
  }, [])

  const liveBalance = wsAccount?.balance ?? status?.balance ?? 10000
  const liveEquity = wsAccount?.equity ?? status?.equity ?? 10000
  const liveDayPnL = wsAccount?.day_pnl ?? status?.day_pnl ?? 0
  const liveDayTrades = wsAccount?.day_trades ?? status?.day_trades ?? 0

  const openPositions = positions.filter((p) => p.status === "open")
  const totalPnL = trades.reduce((sum, t) => sum + t.net_pnl, 0)
  const winTrades = trades.filter((t) => t.net_pnl > 0)
  const winRate = trades.length > 0 ? (winTrades.length / trades.length * 100).toFixed(1) : "0"
  const profitFactor = trades.length > 0
    ? ((winTrades.reduce((s, t) => s + t.net_pnl, 0) / Math.abs(trades.filter(t => t.net_pnl < 0).reduce((s, t) => s + t.net_pnl, 0))) || 0)
    : 0

  const maxDrawdown = account.length > 0
    ? Math.max(...account.map((a, i) => {
        const peak = Math.max(...account.slice(0, i + 1).map(x => x.equity))
        return ((peak - a.equity) / peak) * 100
      }))
    : 0

  const avgHolding = trades.length > 0
    ? (trades.reduce((s, t) => s + t.holding_bars, 0) / trades.length).toFixed(1)
    : "0"

  const chartConfig = {
    equity: { label: "Equity", color: "var(--chart-1)" },
    net_pnl: { label: "Net PnL", color: "var(--chart-1)" },
  } satisfies import("@/components/ui/chart").ChartConfig

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold tracking-tight">Trading Terminal</h1>

      {/* Portfolio Overview */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Balance</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">${liveBalance.toFixed(2)}</div>
            <p className="text-xs text-muted-foreground mt-1">Equity: ${liveEquity.toFixed(2)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total PnL</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className={cn("text-2xl font-bold", totalPnL >= 0 ? "text-emerald-500" : "text-red-500")}>
              ${totalPnL.toFixed(2)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Win Rate</CardTitle>
            <Target className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{winRate}%</div>
            <p className="text-xs text-muted-foreground mt-1">{winTrades.length}/{trades.length} trades</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Max Drawdown</CardTitle>
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-500">{maxDrawdown.toFixed(2)}%</div>
          </CardContent>
        </Card>
      </div>

      {/* Risk Metrics */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground">Profit Factor</p>
            <p className="text-2xl font-bold">{profitFactor.toFixed(2)}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground">Avg Holding Time</p>
            <p className="text-2xl font-bold">{avgHolding} bars</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <p className="text-sm text-muted-foreground">Day PnL</p>
            <p className={cn("text-2xl font-bold", liveDayPnL >= 0 ? "text-emerald-500" : "text-red-500")}>
              ${liveDayPnL.toFixed(2)}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Equity Curve + PnL */}
      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Equity Curve</CardTitle>
          </CardHeader>
          <CardContent className="h-64">
            {account.length === 0 ? (
              <p className="text-muted-foreground text-sm">No data</p>
            ) : (
              <ChartContainer config={chartConfig} className="h-64">
                <AreaChart data={account}>
                  <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                  <XAxis dataKey="ts" tick={{ fontSize: 10 }} hide />
                  <YAxis tick={{ fontSize: 10 }} domain={["auto", "auto"]} />
                  <ChartTooltip content={<ChartTooltipContent />} />
                  <Area type="monotone" dataKey="equity" stroke="var(--color-equity)" fill="var(--color-equity)" fillOpacity={0.1} strokeWidth={2} />
                </AreaChart>
              </ChartContainer>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>PnL per Trade</CardTitle>
          </CardHeader>
          <CardContent className="h-64">
            {trades.length === 0 ? (
              <p className="text-muted-foreground text-sm">No trades yet</p>
            ) : (
              <ChartContainer config={chartConfig} className="h-64">
                <BarChart data={trades.slice(-50).map((t) => ({ ...t, fill: t.net_pnl >= 0 ? "#22c55e" : "#ef4444" }))}>
                  <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                  <XAxis dataKey="id" tick={{ fontSize: 10 }} hide />
                  <YAxis tick={{ fontSize: 10 }} />
                  <ChartTooltip content={<ChartTooltipContent />} />
                  <Bar dataKey="net_pnl" fill="var(--color-net_pnl)" />
                </BarChart>
              </ChartContainer>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Positions + Trades */}
      <Tabs defaultValue="positions">
        <TabsList>
          <TabsTrigger value="positions">Positions ({openPositions.length})</TabsTrigger>
          <TabsTrigger value="trades">Trades ({trades.length})</TabsTrigger>
        </TabsList>

        <TabsContent value="positions">
          <Card>
            <CardHeader>
              <CardTitle>Open Positions</CardTitle>
            </CardHeader>
            <CardContent>
              {openPositions.length === 0 ? (
                <p className="text-sm text-muted-foreground py-4">No open positions</p>
              ) : (
                <div className="space-y-2">
                  {openPositions.map((p) => (
                    <div key={p.id} className="flex items-center justify-between rounded-lg border p-3">
                      <div className="flex items-center gap-3">
                        <Badge variant={p.direction === "long" ? "success" : "destructive"}>
                          {p.direction}
                        </Badge>
                        <span className="font-medium">{p.symbol}</span>
                        <span className="text-sm text-muted-foreground">${p.entry_price.toFixed(2)}</span>
                      </div>
                      <div className="flex items-center gap-4 text-sm">
                        <span>{p.quantity.toFixed(4)}</span>
                        <span className="text-muted-foreground">{p.bars_held} bars</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="trades">
          <Card>
            <CardHeader>
              <CardTitle>Completed Trades</CardTitle>
            </CardHeader>
            <CardContent>
              {trades.length === 0 ? (
                <p className="text-sm text-muted-foreground py-4">No completed trades</p>
              ) : (
                <div className="space-y-2">
                  {trades.map((t) => (
                    <div key={t.id} className="flex items-center justify-between rounded-lg border p-3">
                      <div className="flex items-center gap-3">
                        <Badge variant={t.direction === "long" ? "success" : "destructive"} className="text-xs">
                          {t.direction === "long" ? "L" : "S"}
                        </Badge>
                        <span className="font-medium text-sm">{t.symbol}</span>
                        <span className="text-xs text-muted-foreground">${t.entry_price.toFixed(2)} → ${t.exit_price.toFixed(2)}</span>
                      </div>
                      <div className="flex items-center gap-4">
                        <span className={cn("font-medium", t.net_pnl >= 0 ? "text-emerald-500" : "text-red-500")}>
                          ${t.net_pnl.toFixed(2)}
                        </span>
                        <span className="text-xs text-muted-foreground">{t.holding_bars} bars</span>
                        <span className="text-xs text-muted-foreground">{t.exit_reason}</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
