import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select"
import { api, type Candle, type OrderflowPoint, type Liquidation } from "@/lib/api"
import { useWS } from "@/hooks/useWebSocket"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { LineChart, Line, XAxis, YAxis, CartesianGrid, BarChart, Bar } from "recharts"
import { TrendingUp, Activity, Flame } from "lucide-react"

function cn(...inputs: (string | undefined | false | null)[]) {
  return inputs.filter(Boolean).join(" ")
}

const chartConfig = {
  close: { label: "Close Price", color: "var(--chart-1)" },
  funding_rate: { label: "Funding Rate", color: "var(--chart-1)" },
  oi_value_usd: { label: "Open Interest", color: "var(--chart-2)" },
  ls_ratio: { label: "LS Ratio", color: "var(--chart-3)" },
  volume: { label: "Volume", color: "var(--chart-4)" },
} satisfies import("@/components/ui/chart").ChartConfig

export function MarketData() {
  const [symbol, setSymbol] = useState("BTCUSDT")
  const [tf, setTf] = useState("15m")
  const [candles, setCandles] = useState<Candle[]>([])
  const [orderflow, setOrderflow] = useState<OrderflowPoint[]>([])
  const [liquidations, setLiquidations] = useState<Liquidation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const wsData = useWS("candle")
  const livePrice = symbol === "BTCUSDT" ? wsData?.close : undefined

  const fetchData = async () => {
    try {
      setError(false)
      setLoading(true)
      const [c, o, l] = await Promise.all([
        api.candles(symbol, tf, 150),
        api.orderflow(symbol, 200),
        api.liquidations(symbol, 50),
      ])
      setCandles(c)
      setOrderflow(o)
      setLiquidations(l)
    } catch {
      setError(true)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [symbol, tf])

  const chartData = candles.map((c) => ({
    time: c.time.slice(5, 16),
    close: c.close,
    volume: c.volume,
  }))

  const fundData = orderflow.filter((o) => o.funding_rate != null).slice(-100)
  const oiData = orderflow.filter((o) => o.oi != null).slice(-100)
  const lsData = orderflow.filter((o) => o.buy_ratio != null && o.sell_ratio != null).slice(-50)

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
          <Select value={symbol} onValueChange={setSymbol}>
            <SelectTrigger className="w-[130px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="BTCUSDT">BTC/USDT</SelectItem>
              <SelectItem value="ETHUSDT">ETH/USDT</SelectItem>
              <SelectItem value="SOLUSDT">SOL/USDT</SelectItem>
            </SelectContent>
          </Select>
          <Select value={tf} onValueChange={setTf}>
            <SelectTrigger className="w-[80px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="15m">15m</SelectItem>
              <SelectItem value="1h">1h</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Main Price Chart */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4" /> Close Price — {symbol} {tf}
          </CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? <Skeleton className="h-80 w-full" /> : (
            <ChartContainer config={chartConfig} className="h-80">
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                <XAxis dataKey="time" tick={{ fontSize: 10 }} interval="preserveStartEnd" />
                <YAxis domain={["auto", "auto"]} tick={{ fontSize: 10 }} />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Line type="monotone" dataKey="close" stroke="var(--color-close)" dot={false} strokeWidth={1.5} />
              </LineChart>
            </ChartContainer>
          )}
        </CardContent>
      </Card>

      {/* Orderflow Metrics */}
      <div className="grid gap-6 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-sm">
              <Activity className="h-4 w-4" /> Funding Rate
            </CardTitle>
          </CardHeader>
          <CardContent className="h-48">
            {loading ? <Skeleton className="h-full w-full" /> : (
              <ChartContainer config={chartConfig} className="h-48">
                <LineChart data={fundData}>
                  <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                  <XAxis dataKey="time" tick={{ fontSize: 10 }} hide />
                  <YAxis tick={{ fontSize: 10 }} />
                  <ChartTooltip content={<ChartTooltipContent />} />
                  <Line type="monotone" dataKey="funding_rate" stroke="var(--color-funding_rate)" dot={false} strokeWidth={1.5} />
                </LineChart>
              </ChartContainer>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-sm">
              <Activity className="h-4 w-4" /> Open Interest
            </CardTitle>
          </CardHeader>
          <CardContent className="h-48">
            {loading ? <Skeleton className="h-full w-full" /> : (
              <ChartContainer config={chartConfig} className="h-48">
                <LineChart data={oiData}>
                  <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                  <XAxis dataKey="time" tick={{ fontSize: 10 }} hide />
                  <YAxis tick={{ fontSize: 10 }} />
                  <ChartTooltip content={<ChartTooltipContent />} />
                  <Line type="monotone" dataKey="oi_value_usd" stroke="var(--color-oi_value_usd)" dot={false} strokeWidth={1.5} />
                </LineChart>
              </ChartContainer>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-sm">
              <Activity className="h-4 w-4" /> L/S Ratio
            </CardTitle>
          </CardHeader>
          <CardContent className="h-48">
            {loading ? <Skeleton className="h-full w-full" /> : (
              <ChartContainer config={chartConfig} className="h-48">
                <LineChart data={lsData}>
                  <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                  <XAxis dataKey="time" tick={{ fontSize: 10 }} hide />
                  <YAxis tick={{ fontSize: 10 }} />
                  <ChartTooltip content={<ChartTooltipContent />} />
                  <Line type="monotone" dataKey="buy_ratio" stroke="var(--color-ls_ratio)" dot={false} strokeWidth={1.5} />
                </LineChart>
              </ChartContainer>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Liquidations */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Flame className="h-4 w-4" /> Recent Liquidations
          </CardTitle>
        </CardHeader>
        <CardContent>
          {liquidations.length === 0 ? (
            <p className="text-sm text-muted-foreground">No liquidation events</p>
          ) : (
            <div className="space-y-2">
              {liquidations.slice(0, 10).map((liq, i) => (
                <div key={i} className="flex items-center justify-between rounded-lg border p-2">
                  <div className="flex items-center gap-2">
                    <Badge variant={liq.side === "Buy" ? "success" : "destructive"} className="text-xs">
                      {liq.side}
                    </Badge>
                    <span className="text-sm font-mono">{liq.size.toFixed(4)}</span>
                  </div>
                  <div className="flex items-center gap-4 text-xs text-muted-foreground">
                    <span>${liq.price.toFixed(2)}</span>
                    <span>${liq.value_usd.toLocaleString()}</span>
                    <span>{liq.time.slice(5, 16)}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
