import type { ReactNode } from "react"
import { Navigate, Route, Routes } from "react-router-dom"
import { Layout } from "@/components/Layout"
import { getToken } from "@/lib/api"
import { AnalyzePage } from "@/pages/Analyze"
import { AuthPage } from "@/pages/Auth"
import { BriefPage } from "@/pages/Brief"
import { ProjectFormPage } from "@/pages/ProjectForm"
import { ProjectsPage } from "@/pages/Projects"
import { StrategyPage } from "@/pages/Strategy"

function Private({ children }: { children: ReactNode }) {
  if (!getToken()) return <Navigate to="/login" replace />
  return children
}

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Navigate to={getToken() ? "/projects" : "/login"} replace />} />
        <Route path="/login" element={<AuthPage mode="login" />} />
        <Route path="/register" element={<AuthPage mode="register" />} />
        <Route path="/projects" element={<Private><ProjectsPage /></Private>} />
        <Route path="/projects/new" element={<Private><ProjectFormPage /></Private>} />
        <Route path="/projects/:id" element={<Private><BriefPage /></Private>} />
        <Route path="/projects/:id/analyze/:jobId" element={<Private><AnalyzePage /></Private>} />
        <Route path="/projects/:id/strategy" element={<Private><StrategyPage /></Private>} />
      </Route>
    </Routes>
  )
}
