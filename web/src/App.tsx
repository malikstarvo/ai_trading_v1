import { useState } from "react"
import { Sidebar } from "@/components/layout/Sidebar"
import { Dashboard } from "@/pages/Dashboard"
import { Market } from "@/pages/Market"
import { Features } from "@/pages/Features"
import { Trading } from "@/pages/Trading"
import { Model } from "@/pages/Model"
import { System } from "@/pages/System"

const pages: Record<string, React.FC> = {
  "/": Dashboard,
  "/market": Market,
  "/features": Features,
  "/trading": Trading,
  "/model": Model,
  "/system": System,
}

export default function App() {
  const [page, setPage] = useState("/")
  const PageComponent = pages[page] || Dashboard

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar current={page} onNavigate={setPage} />
      <main className="flex-1 overflow-y-auto bg-background">
        <div className="mx-auto max-w-7xl p-6">
          <PageComponent />
        </div>
      </main>
    </div>
  )
}
