import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Skeleton } from "@/components/ui/skeleton"
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select"
import { api, type FeatureNaN, type FeatureRanking, type FeatureRow } from "@/lib/api"
import { BarChart, Bar, XAxis, YAxis, CartesianGrid } from "recharts"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"

function NaNBar({ pct }: { pct: number }) {
  const color = pct > 80 ? "#ef4444" : pct > 40 ? "#f59e0b" : "#22c55e"
  return (
    <div className="flex items-center gap-2">
      <div className="h-2 w-24 rounded-full bg-muted overflow-hidden">
        <div className="h-full rounded-full transition-all" style={{ width: `${pct}%`, backgroundColor: color }} />
      </div>
      <span className="text-xs font-medium">{pct.toFixed(1)}%</span>
    </div>
  )
}

export function Features() {
  const [symbol, setSymbol] = useState("BTCUSDT")
  const [tf, setTf] = useState("15m")
  const [nanData, setNanData] = useState<FeatureNaN[]>([])
  const [ranking, setRanking] = useState<FeatureRanking[]>([])
  const [latest, setLatest] = useState<FeatureRow | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    Promise.all([
      api.featuresNaN(symbol, tf),
      api.featuresRanking(),
      api.featuresLatest(symbol, tf),
    ]).then(([nan, rank, lat]) => {
      setNanData(nan)
      setRanking(rank)
      setLatest(lat)
    }).catch(console.error).finally(() => setLoading(false))
  }, [symbol, tf])

  if (loading) return (
    <div className="space-y-4">
      <h1 className="text-3xl font-bold tracking-tight">Features</h1>
      <Skeleton className="h-96 w-full" />
    </div>
  )

  const top10 = ranking.slice(0, 10)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Features Analysis</h1>
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

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Feature Ranking (Edge Study)</CardTitle>
          </CardHeader>
          <CardContent className="h-72">
            <ChartContainer config={{ score: { label: "Edge Score", color: "var(--chart-1)" } }} className="h-72">
              <BarChart data={top10} layout="vertical" margin={{ left: 100 }}>
                <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                <XAxis type="number" domain={[0, 1]} tick={{ fontSize: 10 }} />
                <YAxis dataKey="feature" type="category" tick={{ fontSize: 10 }} width={90} />
                <ChartTooltip content={<ChartTooltipContent />} />
                <Bar dataKey="score" fill="var(--color-score)" radius={[0, 4, 4, 0]} />
              </BarChart>
            </ChartContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Latest Feature Values</CardTitle>
          </CardHeader>
          <CardContent className="max-h-72 overflow-y-auto">
            {latest ? (
              <div className="space-y-1 text-sm">
                {Object.entries(latest).filter(([k]) => k !== "ts" && k !== "feature_set_id").map(([k, v]) => (
                  <div key={k} className="flex justify-between border-b py-1">
                    <span className="text-muted-foreground">{k}</span>
                    <span className="font-mono">{typeof v === "number" ? v.toFixed(4) : String(v ?? "NaN")}</span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No data</p>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>NaN Coverage — {symbol} {tf}</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Feature</TableHead>
                <TableHead>Non-Null</TableHead>
                <TableHead>Total</TableHead>
                <TableHead>NaN %</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {nanData.map((row) => (
                <TableRow key={row.column}>
                  <TableCell className="font-mono text-xs">{row.column}</TableCell>
                  <TableCell>{row.non_null.toLocaleString()}</TableCell>
                  <TableCell>{row.total_rows.toLocaleString()}</TableCell>
                  <TableCell>
                    <NaNBar pct={row.nan_pct} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
