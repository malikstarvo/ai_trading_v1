import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { api, type PaperStatus, type Position, type Trade, type AccountSnapshot } from "@/lib/api"
import { useWS } from "@/hooks/useWebSocket"
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, AreaChart, Area } from "recharts"
import { TrendingUp, DollarSign, BarChart3 } from "lucide-react"

export function Trading() {
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

  // Merge live WS account data
  const liveBalance = wsAccount?.balance ?? status?.balance ?? 10000
  const liveEquity = wsAccount?.equity ?? status?.equity ?? 10000
  const liveDayPnL = wsAccount?.day_pnl ?? status?.day_pnl ?? 0
  const liveDayTrades = wsAccount?.day_trades ?? status?.day_trades ?? 0

  if (loading) return (
    <div className="space-y-4">
      <h1 className="text-3xl font-bold tracking-tight">Paper Trading</h1>
      <Skeleton className="h-64 w-full" />
    </div>
  )

  const openPositions = positions.filter((p) => p.status === "open")
  const totalPnL = trades.reduce((sum, t) => sum + t.net_pnl, 0)
  const winTrades = trades.filter((t) => t.net_pnl > 0)
  const winRate = trades.length > 0 ? (winTrades.length / trades.length * 100).toFixed(1) : "0"

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold tracking-tight">Paper Trading</h1>

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
            <div className={`text-2xl font-bold ${totalPnL >= 0 ? "text-emerald-500" : "text-red-500"}`}>
              ${totalPnL.toFixed(2)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Win Rate</CardTitle>
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{winRate}%</div>
            <p className="text-xs text-muted-foreground mt-1">{winTrades.length}/{trades.length} trades</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Day Trades</CardTitle>
            <BarChart3 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{liveDayTrades}</div>
            <p className={`text-xs mt-1 ${liveDayPnL >= 0 ? "text-emerald-500" : "text-red-500"}`}>
              Day PnL: ${liveDayPnL.toFixed(2)}
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Equity Curve</CardTitle>
          </CardHeader>
          <CardContent className="h-64">
            {account.length === 0 ? (
              <p className="text-muted-foreground text-sm">No data</p>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={account}>
                  <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                  <XAxis dataKey="ts" tick={{ fontSize: 10 }} hide />
                  <YAxis tick={{ fontSize: 10 }} domain={["auto", "auto"]} />
                  <Tooltip />
                  <Area type="monotone" dataKey="equity" stroke="#6366f1" fill="#6366f1" fillOpacity={0.1} strokeWidth={2} />
                </AreaChart>
              </ResponsiveContainer>
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
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={trades.slice(-50).map((t) => ({ ...t, fill: t.net_pnl >= 0 ? "#22c55e" : "#ef4444" }))}>
                  <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                  <XAxis dataKey="id" tick={{ fontSize: 10 }} hide />
                  <YAxis tick={{ fontSize: 10 }} />
                  <Tooltip />
                  <Bar dataKey="net_pnl" fill="#6366f1" />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>

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
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Symbol</TableHead>
                      <TableHead>Direction</TableHead>
                      <TableHead>Entry</TableHead>
                      <TableHead>Size</TableHead>
                      <TableHead>Bars</TableHead>
                      <TableHead>Open</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {openPositions.map((p) => (
                      <TableRow key={p.id}>
                        <TableCell>{p.symbol}</TableCell>
                        <TableCell>
                          <Badge variant={p.direction === "long" ? "success" : "destructive"}>
                            {p.direction}
                          </Badge>
                        </TableCell>
                        <TableCell>${p.entry_price.toFixed(2)}</TableCell>
                        <TableCell>{p.quantity.toFixed(4)}</TableCell>
                        <TableCell>{p.bars_held}</TableCell>
                        <TableCell className="text-xs">{new Date(p.open_ts).toLocaleString()}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
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
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Symbol</TableHead>
                      <TableHead>Dir</TableHead>
                      <TableHead>Entry</TableHead>
                      <TableHead>Exit</TableHead>
                      <TableHead>PnL</TableHead>
                      <TableHead>Return</TableHead>
                      <TableHead>Bars</TableHead>
                      <TableHead>Exit Reason</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {trades.map((t) => (
                      <TableRow key={t.id}>
                        <TableCell>{t.symbol}</TableCell>
                        <TableCell>
                          <Badge variant={t.direction === "long" ? "success" : "destructive"}>
                            {t.direction === "long" ? "L" : "S"}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-xs">${t.entry_price.toFixed(2)}</TableCell>
                        <TableCell className="text-xs">${t.exit_price.toFixed(2)}</TableCell>
                        <TableCell className={`font-medium ${t.net_pnl >= 0 ? "text-emerald-500" : "text-red-500"}`}>
                          ${t.net_pnl.toFixed(2)}
                        </TableCell>
                        <TableCell>{(t.return_pct * 100).toFixed(2)}%</TableCell>
                        <TableCell>{t.holding_bars}</TableCell>
                        <TableCell className="text-xs">{t.exit_reason}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
