import { cn } from "@/lib/utils"
import { createContext, useContext, useState, useCallback } from "react"

interface TabsCtx {
  value: string
  setValue: (v: string) => void
}
const Ctx = createContext<TabsCtx>({ value: "", setValue: () => {} })

function Tabs({ defaultValue, className, children }: { defaultValue: string; className?: string; children: React.ReactNode }) {
  const [value, setValue] = useState(defaultValue)
  return (
    <Ctx.Provider value={{ value, setValue }}>
      <div className={cn("", className)}>{children}</div>
    </Ctx.Provider>
  )
}

function TabsList({ className, children }: { className?: string; children: React.ReactNode }) {
  return (
    <div className={cn("inline-flex h-9 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground", className)}>
      {children}
    </div>
  )
}

function TabsTrigger({ value, className, children }: { value: string; className?: string; children: React.ReactNode }) {
  const { value: current, setValue } = useContext(Ctx)
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center whitespace-nowrap rounded-md px-3 py-1 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
        current === value && "bg-background text-foreground shadow",
        className
      )}
      data-state={current === value ? "active" : "inactive"}
      onClick={() => setValue(value)}
    >
      {children}
    </button>
  )
}

function TabsContent({ value, className, children }: { value: string; className?: string; children: React.ReactNode }) {
  const { value: current } = useContext(Ctx)
  if (current !== value) return null
  return (
    <div className={cn("mt-2 ring-offset-background", className)} data-state="active">
      {children}
    </div>
  )
}

export { Tabs, TabsList, TabsTrigger, TabsContent }
