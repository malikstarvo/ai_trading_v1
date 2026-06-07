import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { api, type ModelStatus, type SystemInfo } from "@/lib/api"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { BarChart, Bar, XAxis, YAxis, CartesianGrid } from "recharts"
import { BrainCircuit, RefreshCw, Play, Clock, Server, Activity } from "lucide-react"

function cn(...inputs: (string | undefined | false | null)[]) {
  return inputs.filter(Boolean).join(" ")
}

export function SystemModels() {
  const [modelStatus, setModelStatus] = useState<ModelStatus | null>(null)
  const [sys, setSys] = useState<SystemInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  const fetchData = () => {
    setLoading(true)
    Promise.all([
      api.modelStatus(),
      api.system(),
    ]).then(([m, s]) => {
      setModelStatus(m)
      setSys(s)
    }).catch(console.error).finally(() => setLoading(false))
  }

  useEffect(() => { fetchData() }, [])

  const runAction = async (action: string) => {
    setActionLoading(action)
    try {
      await fetch(`/api/actions/${action}`, { method: "POST" })
    } catch (err) {
      console.error(err)
    } finally {
      setActionLoading(null)
    }
  }

  const models = modelStatus?.models ?? []
  const aucData = models.map((m) => ({
    name: `${m.horizon}bar`,
    auc: m.auc ?? 0,
    exists: m.exists,
  }))

  const chartConfig = {
    auc: { label: "Test AUC", color: "var(--chart-1)" },
  } satisfies import("@/components/ui/chart").ChartConfig

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">System & Models</h1>
        <Button variant="outline" size="sm" onClick={fetchData}>
          <RefreshCw className="h-4 w-4 mr-1" /> Refresh
        </Button>
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Play className="h-4 w-4" /> Quick Actions
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-3">
            <Button
              onClick={() => runAction("feature-backfill")}
              disabled={actionLoading === "feature-backfill"}
              variant="outline"
            >
              {actionLoading === "feature-backfill" ? (
                <RefreshCw className="h-4 w-4 mr-1 animate-spin" />
              ) : (
                <Activity className="h-4 w-4 mr-1" />
              )}
              Feature Backfill
            </Button>
            <Button
              onClick={() => runAction("edge-study")}
              disabled={actionLoading === "edge-study"}
              variant="outline"
            >
              {actionLoading === "edge-study" ? (
                <RefreshCw className="h-4 w-4 mr-1 animate-spin" />
              ) : (
                <BrainCircuit className="h-4 w-4 mr-1" />
              )}
              Edge Study
            </Button>
            <Button
              onClick={() => runAction("model-training")}
              disabled={actionLoading === "model-training"}
              variant="outline"
            >
              {actionLoading === "model-training" ? (
                <RefreshCw className="h-4 w-4 mr-1 animate-spin" />
              ) : (
                <BrainCircuit className="h-4 w-4 mr-1" />
              )}
              Model Training
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Pipeline Status */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">Collector</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2 mb-2">
              <div className={cn("h-2.5 w-2.5 rounded-full", sys?.collector.ws_connected ? "bg-emerald-500" : "bg-red-500")} />
              <span className="text-sm font-medium">{sys?.collector.ws_connected ? "Connected" : "Disconnected"}</span>
            </div>
            <p className="text-xs text-muted-foreground">
              Uptime: {sys?.collector.uptime_hours.toFixed(1) ?? 0}h
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">Database</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{sys?.db.size_mb.toFixed(1) ?? 0} MB</div>
            <p className="text-xs text-muted-foreground mt-1">{sys?.db.tables.length ?? 0} tables</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">Paper Trader</CardTitle>
          </CardHeader>
          <CardContent>
            <Badge variant={sys?.paper_trader.pid ? "success" : "warning"}>
              {sys?.paper_trader.pid ? `Running (PID ${sys.paper_trader.pid})` : "Not running"}
            </Badge>
          </CardContent>
        </Card>
      </div>

      {/* Model Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        {models.map((m) => (
          <Card key={m.horizon}>
            <CardHeader>
              <CardTitle className="flex items-center justify-between text-base">
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

      {/* Model Performance Chart */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <BrainCircuit className="h-4 w-4" /> Model Performance
          </CardTitle>
        </CardHeader>
        <CardContent className="h-72">
          {aucData.length === 0 ? (
            <p className="text-muted-foreground text-sm">No trained models yet</p>
          ) : (
            <ChartContainer config={chartConfig} className="h-72">
              <BarChart data={aucData}>
                <CartesianGrid strokeDasharray="3 3" opacity={0.2} />
                <XAxis dataKey="name" />
                <YAxis domain={[0, 1]} tick={{ fontSize: 10 }} />
                <ChartTooltip content={<ChartTooltipContent formatter={(v: number) => [`${(v * 100).toFixed(1)}%`, "AUC"]} />} />
                <Bar dataKey="auc" fill="var(--color-auc)" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ChartContainer>
          )}
        </CardContent>
      </Card>

      {/* Cron Schedule */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Clock className="h-4 w-4" /> Cron Schedule
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {(sys?.crons ?? []).map((c: { schedule: string; command: string }, i: number) => (
              <div key={i} className="flex items-center justify-between rounded-lg border p-3">
                <div className="flex items-center gap-3">
                  <code className="rounded bg-muted px-2 py-1 text-xs font-mono">{c.schedule}</code>
                  <span className="text-sm text-muted-foreground">{c.command}</span>
                </div>
                <Badge variant="outline" className="text-xs">Scheduled</Badge>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
