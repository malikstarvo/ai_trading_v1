import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Skeleton } from "@/components/ui/skeleton"
import { api, type SystemInfo } from "@/lib/api"
import { Server, Database, Clock, Activity } from "lucide-react"

export function System() {
  const [sys, setSys] = useState<SystemInfo | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.system().then(setSys).catch(console.error).finally(() => setLoading(false))
  }, [])

  if (loading) return (
    <div className="space-y-4">
      <h1 className="text-3xl font-bold tracking-tight">System</h1>
      <Skeleton className="h-64 w-full" />
    </div>
  )

  const crons = sys?.crons ?? []
  const tables = sys?.db?.tables ?? []

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold tracking-tight">System Status</h1>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Collector</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2 mb-2">
              <div className={`h-2.5 w-2.5 rounded-full ${sys?.collector.ws_connected ? "bg-emerald-500" : "bg-red-500"}`} />
              <span className="text-sm font-medium">{sys?.collector.ws_connected ? "Connected" : "Disconnected"}</span>
            </div>
            <p className="text-xs text-muted-foreground">
              Uptime: {sys?.collector.uptime_hours.toFixed(1) ?? 0}h<br />
              Last heartbeat: {sys?.collector.last_heartbeat ? new Date(sys.collector.last_heartbeat).toLocaleString() : "N/A"}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Database</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{sys?.db.size_mb.toFixed(1) ?? 0} MB</div>
            <p className="text-xs text-muted-foreground mt-1">{tables.length} tables tracked</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Paper Trader</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <Badge variant={sys?.paper_trader.pid ? "success" : "warning"}>
              {sys?.paper_trader.pid ? `Running (PID ${sys.paper_trader.pid})` : "Not running"}
            </Badge>
            {sys?.paper_trader.uptime_hours ? (
              <p className="text-xs text-muted-foreground mt-2">
                Uptime: {sys.paper_trader.uptime_hours.toFixed(1)}h
              </p>
            ) : null}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Clock className="h-4 w-4" /> Cron Schedule
            </CardTitle>
          </CardHeader>
          <CardContent>
            {crons.length === 0 ? (
              <p className="text-sm text-muted-foreground">No cron jobs configured</p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Schedule</TableHead>
                    <TableHead>Command</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {crons.map((c: { schedule: string; command: string }, i: number) => (
                    <TableRow key={i}>
                      <TableCell><code className="rounded bg-muted px-1.5 py-0.5 text-xs">{c.schedule}</code></TableCell>
                      <TableCell className="font-mono text-xs">{c.command}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Database className="h-4 w-4" /> Table Row Counts
            </CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Table</TableHead>
                  <TableHead className="text-right">Rows</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tables.map((t: { name: string; count: number }) => (
                    <TableRow key={t.name}>
                      <TableCell className="font-mono text-xs">{t.name}</TableCell>
                      <TableCell className="text-right">{t.count.toLocaleString()}</TableCell>
                    </TableRow>
                  ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
