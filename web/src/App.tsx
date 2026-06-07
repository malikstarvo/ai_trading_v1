import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom"
import { Layout } from "@/components/layout/Layout"
import { Overview } from "@/pages/Overview"
import { MarketData } from "@/pages/MarketData"
import { DataQuality } from "@/pages/DataQuality"
import { TradingTerminal } from "@/pages/TradingTerminal"
import { SystemModels } from "@/pages/SystemModels"

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Navigate to="/overview" replace />} />
          <Route path="/overview" element={<Overview />} />
          <Route path="/market" element={<MarketData />} />
          <Route path="/data" element={<DataQuality />} />
          <Route path="/trading" element={<TradingTerminal />} />
          <Route path="/system" element={<SystemModels />} />
          <Route path="*" element={<Navigate to="/overview" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
