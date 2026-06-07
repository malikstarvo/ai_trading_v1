import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Badge } from "@/components/ui/badge"
import { api, type Candle, type OrderflowPoint } from "@/lib/api"
import { useWS } from "@/hooks/useWebSocket"
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from "recharts"

function CloseChart({ data }: { data: Candle[] }) {
  if (data.length === 0) return <div className="flex items-center justify-center h-64 text-muted-foreground">No candle data</div>

  const chartData = data.map((c) => ({
    time: c.time.slice(5, 16),
    close: c.close,
    volume: c.volume,
  }))

  return (
    <div className="h-80">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
          <XAxis dataKey="time" tick={{ fontSize: 10 }} interval="preserveStartEnd" />
          <YAxis domain={["auto", "auto"]} tick={{ fontSize: 10 }} />
          <Tooltip />
          <Line type="monotone" dataKey="close" stroke="#6366f1" dot={false} strokeWidth={1.5} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}

export function Market() {
  const [symbol, setSymbol] = useState("BTCUSDT")
  const [tf, setTf] = useState("15m")
  const [candles, setCandles] = useState<Candle[]>([])
  const [orderflow, setOrderflow] = useState<OrderflowPoint[]>([])
  const [loading, setLoading] = useState(true)
  const wsData = useWS("candle")

  // Append live candle to chart if symbol matches
  const livePrice = symbol === "BTCUSDT" ? wsData?.close : undefined

  useEffect(() => {
    setLoading(true)
    Promise.all([
      api.candles(symbol, tf, 150),
      api.orderflow(symbol, 200),
    ]).then(([c, o]) => {
      setCandles(c)
      setOrderflow(o)
    }).catch(console.error).finally(() => setLoading(false))
  }, [symbol, tf])

  const fundData = orderflow.filter((o) => o.funding_rate != null).slice(-100)
  const oiData = orderflow.filter((o) => o.oi != null).slice(-100)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-3xl font-bold tracking-tight">Market Data</h1>
          {livePrice && (
            <Badge variant="outline" className="text-base font-mono px-3 py-1">
              ${livePrice.toFixed(2)}
            </Badge>
          )}
        </div>
        <div className="flex gap-2">
          <select className="rounded-lg border bg-background px-3 py-1.5 text-sm" value={symbol} onChange={(e) => setSymbol(e.target.value)}>
            <option value="BTCUSDT">BTC/USDT</option>
            <option value="ETHUSDT">ETH/USDT</option>
            <option value="SOLUSDT">SOL/USDT</option>
          </select>
          <select className="rounded-lg border bg-background px-3 py-1.5 text-sm" value={tf} onChange={(e) => setTf(e.target.value)}>
            <option value="15m">15m</option>
            <option value="1h">1h</option>
          </select>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Close Price — {symbol} {tf}</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? <Skeleton className="h-80 w-full" /> : <CloseChart data={candles} />}
        </CardContent>
      </Card>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Funding Rate</CardTitle>
          </CardHeader>
          <CardContent className="h-48">
            {loading ? <Skeleton className="h-full w-full" /> : (
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={fundData}>
                  <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                  <XAxis dataKey="time" tick={{ fontSize: 10 }} hide />
                  <YAxis tick={{ fontSize: 10 }} />
                  <Tooltip />
                  <Line type="monotone" dataKey="funding_rate" stroke="#6366f1" dot={false} strokeWidth={1.5} />
                </LineChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Open Interest</CardTitle>
          </CardHeader>
          <CardContent className="h-48">
            {loading ? <Skeleton className="h-full w-full" /> : (
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={oiData}>
                  <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                  <XAxis dataKey="time" tick={{ fontSize: 10 }} hide />
                  <YAxis tick={{ fontSize: 10 }} />
                  <Tooltip />
                  <Line type="monotone" dataKey="oi_value_usd" stroke="#22c55e" dot={false} strokeWidth={1.5} />
                </LineChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
