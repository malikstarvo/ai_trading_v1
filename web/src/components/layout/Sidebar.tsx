import { cn } from "@/lib/utils"
import {
  LayoutDashboard, CandlestickChart, Activity, TrendingUp,
  BrainCircuit, Server, Moon, Sun,
} from "lucide-react"
import { useState } from "react"

const nav = [
  { href: "/", label: "Dashboard", icon: LayoutDashboard },
  { href: "/market", label: "Market", icon: CandlestickChart },
  { href: "/features", label: "Features", icon: Activity },
  { href: "/trading", label: "Trading", icon: TrendingUp },
  { href: "/model", label: "Model", icon: BrainCircuit },
  { href: "/system", label: "System", icon: Server },
]

export function Sidebar({ current, onNavigate }: { current: string; onNavigate: (path: string) => void }) {
  const [collapsed, setCollapsed] = useState(false)
  const [dark, setDark] = useState(() => document.documentElement.classList.contains("dark"))

  const toggleTheme = () => {
    const next = !dark
    setDark(next)
    document.documentElement.classList.toggle("dark", next)
  }

  return (
    <aside className={cn(
      "flex flex-col border-r bg-sidebar-background text-sidebar-foreground transition-all duration-200",
      collapsed ? "w-16" : "w-56"
    )}>
      <div className="flex h-14 items-center gap-2 border-b px-4">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground text-sm font-bold">
          AT
        </div>
        {!collapsed && <span className="font-semibold">AI Trading</span>}
      </div>

      <nav className="flex-1 space-y-1 p-2">
        {nav.map((item) => (
          <button
            key={item.href}
            onClick={() => onNavigate(item.href)}
            className={cn(
              "flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
              current === item.href
                ? "bg-sidebar-accent text-sidebar-accent-foreground"
                : "hover:bg-sidebar-accent/50 text-sidebar-foreground/70 hover:text-sidebar-foreground"
            )}
          >
            <item.icon className="h-4 w-4 shrink-0" />
            {!collapsed && item.label}
          </button>
        ))}
      </nav>

      <div className="border-t p-2 space-y-1">
        <button
          onClick={toggleTheme}
          className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium hover:bg-sidebar-accent/50 text-sidebar-foreground/70 hover:text-sidebar-foreground transition-colors"
        >
          {dark ? <Sun className="h-4 w-4 shrink-0" /> : <Moon className="h-4 w-4 shrink-0" />}
          {!collapsed && (dark ? "Light" : "Dark")}
        </button>
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium hover:bg-sidebar-accent/50 text-sidebar-foreground/70 hover:text-sidebar-foreground transition-colors"
        >
          <div className="h-4 w-4 shrink-0 flex items-center justify-center">
            {collapsed ? "→" : "←"}
          </div>
          {!collapsed && "Collapse"}
        </button>
      </div>
    </aside>
  )
}
