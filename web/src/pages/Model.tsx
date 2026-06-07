import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { api, type ModelStatus } from "@/lib/api"
import { BarChart, Bar, XAxis, YAxis, CartesianGrid } from "recharts"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { BrainCircuit, RefreshCw } from "lucide-react"

function fmtAUC(v: number | undefined): string {
  if (v === undefined) return "N/A"
  return (v * 100).toFixed(1) + "%"
}

export function Model() {
  const [modelStatus, setModelStatus] = useState<ModelStatus | null>(null)
  const [loading, setLoading] = useState(true)

  const fetchModel = () => {
    setLoading(true)
    api.modelStatus()
      .then(setModelStatus)
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchModel() }, [])

  if (loading) return (
    <div className="space-y-4">
      <h1 className="text-3xl font-bold tracking-tight">Model Training</h1>
      <Skeleton className="h-64 w-full" />
    </div>
  )

  const models = modelStatus?.models ?? []
  const aucData = models.map((m) => ({
    name: `${m.horizon}bar`,
    auc: m.auc ?? 0,
    exists: m.exists,
  }))

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Model Training</h1>
        <Button variant="outline" size="sm" onClick={fetchModel}>
          <RefreshCw className="h-4 w-4 mr-1" /> Refresh
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        {models.map((m) => (
          <Card key={m.horizon}>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>{m.horizon}-Bar Model</span>
                {m.exists ? (
                  <Badge variant="success">Trained</Badge>
                ) : (
                  <Badge variant="warning">Not trained</Badge>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {m.exists ? (
                <div className="space-y-2">
                  <div className="text-3xl font-bold">{m.auc ? (m.auc * 100).toFixed(1) : "N/A"}%</div>
                  <p className="text-xs text-muted-foreground">Test AUC</p>
                  <p className="text-xs text-muted-foreground">
                    Trained: {m.trained_at ? new Date(m.trained_at).toLocaleString() : "Unknown"}
                  </p>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">Run weekly retrain to train this model</p>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <BrainCircuit className="h-4 w-4" /> Model Performance
          </CardTitle>
        </CardHeader>
          <CardContent className="h-72">
          {aucData.length === 0 ? (
            <p className="text-muted-foreground text-sm">No trained models yet. Models are retrained weekly via cron (Sun 4AM UTC).</p>
          ) : (
            <ChartContainer config={{ auc: { label: "Test AUC", color: "var(--chart-1)" } }} className="h-72">
              <BarChart data={aucData}>
                <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                <XAxis dataKey="name" />
                <YAxis domain={[0, 1]} tick={{ fontSize: 10 }} />
                <ChartTooltip content={<ChartTooltipContent formatter={(v: number) => [fmtAUC(v), "AUC"]} />} />
                <Bar dataKey="auc" fill="var(--color-auc)" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ChartContainer>
          )}
          </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Training Pipeline</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-sm">
          <div className="flex items-center justify-between border-b pb-2">
            <span>Symbol</span>
            <span className="font-medium">{modelStatus?.symbol ?? "BTCUSDT"}</span>
          </div>
          <div className="flex items-center justify-between border-b pb-2">
            <span>Timeframe</span>
            <span className="font-medium">{modelStatus?.timeframe ?? "15m"}</span>
          </div>
          <div className="flex items-center justify-between border-b pb-2">
            <span>Schedule</span>
            <span className="font-medium">Every Sunday 4:00 AM UTC</span>
          </div>
          <div className="flex items-center justify-between">
            <span>Script</span>
            <span className="font-mono text-xs">scripts/run-xgb-retrain.sh</span>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
