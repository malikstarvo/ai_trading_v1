import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select"
import { api, type FeatureNaN, type LabelDist, type DataOverview } from "@/lib/api"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { BarChart, Bar, XAxis, YAxis, CartesianGrid } from "recharts"
import { Shield, AlertTriangle, CircleCheck } from "lucide-react"

function cn(...inputs: (string | undefined | false | null)[]) {
  return inputs.filter(Boolean).join(" ")
}

export function DataQuality() {
  const [symbol, setSymbol] = useState("BTCUSDT")
  const [tf, setTf] = useState("15m")
  const [nanData, setNanData] = useState<FeatureNaN[]>([])
  const [labels, setLabels] = useState<LabelDist[]>([])
  const [dataOverview, setDataOverview] = useState<DataOverview | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    Promise.all([
      api.featuresNaN(symbol, tf),
      api.labels(symbol, tf),
      api.dataOverview(),
    ]).then(([nan, lab, overview]) => {
      setNanData(nan)
      setLabels(lab)
      setDataOverview(overview)
    }).catch(console.error).finally(() => setLoading(false))
  }, [symbol, tf])

  const symbolData = dataOverview?.symbols.find(s => s.symbol === symbol && s.timeframe === tf)
  const qualityScore = symbolData ? (
    (100 - (parseFloat(symbolData.oi_nan_pct || "0"))) * 0.5 +
    (100 - (parseFloat(symbolData.ls_nan_pct || "0"))) * 0.5
  ) : 0

  const chartConfig = {
    success_rate: { label: "Success Rate", color: "var(--chart-1)" },
  } satisfies import("@/components/ui/chart").ChartConfig

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Data Quality</h1>
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

      {/* Quality Score */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Feature Quality Score</p>
                <p className="text-3xl font-bold">{qualityScore.toFixed(1)}%</p>
              </div>
              {qualityScore >= 90 ? (
                <CircleCheck className="h-8 w-8 text-emerald-500" />
              ) : qualityScore >= 50 ? (
                <AlertTriangle className="h-8 w-8 text-amber-500" />
              ) : (
                <AlertTriangle className="h-8 w-8 text-red-500" />
              )}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">OI NaN Rate</p>
                <p className={cn("text-3xl font-bold", parseFloat(symbolData?.oi_nan_pct || "0") > 50 ? "text-red-500" : "text-emerald-500")}>
                  {symbolData?.oi_nan_pct || "0"}%
                </p>
              </div>
              <Shield className="h-8 w-8 text-muted-foreground" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">LS Ratio NaN</p>
                <p className="text-3xl font-bold text-emerald-500">
                  {symbolData?.ls_nan_pct || "0"}%
                </p>
              </div>
              <Shield className="h-8 w-8 text-muted-foreground" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Label Distribution */}
      <Card>
        <CardHeader>
          <CardTitle>Label Distribution by Horizon</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? <Skeleton className="h-64 w-full" /> : labels.length === 0 ? (
            <p className="text-sm text-muted-foreground">No labels available</p>
          ) : (
            <ChartContainer config={chartConfig} className="h-64">
              <BarChart data={labels.map(l => ({ ...l, name: `${l.horizon}bar` }))}>
                <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                <XAxis dataKey="name" />
                <YAxis domain={[0, 100]} tick={{ fontSize: 10 }} />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar dataKey="success_rate" fill="var(--color-success_rate)" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ChartContainer>
          )}
        </CardContent>
      </Card>

      {/* NaN Coverage */}
      <Card>
        <CardHeader>
          <CardTitle>NaN Coverage — {symbol} {tf}</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? <Skeleton className="h-96 w-full" /> : (
            <div className="space-y-3">
              {nanData.map((row) => {
                const pct = row.nan_pct
                const color = pct > 80 ? "bg-red-500" : pct > 40 ? "bg-amber-500" : "bg-emerald-500"
                return (
                  <div key={row.column} className="flex items-center gap-4">
                    <span className="w-32 font-mono text-xs text-muted-foreground">{row.column}</span>
                    <div className="flex-1 h-2 rounded-full bg-muted overflow-hidden">
                      <div className={cn("h-full rounded-full", color)} style={{ width: `${pct}%` }} />
                    </div>
                    <span className="w-12 text-right text-xs font-medium">{pct.toFixed(1)}%</span>
                    <span className="w-16 text-right text-xs text-muted-foreground">{row.non_null.toLocaleString()}</span>
                  </div>
                )
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
