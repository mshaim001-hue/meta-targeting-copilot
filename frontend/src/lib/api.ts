import type { Job, Project, StrategyDoc } from "./types"

const TOKEN_KEY = "copilot_token"

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string | null) {
  if (token) localStorage.setItem(TOKEN_KEY, token)
  else localStorage.removeItem(TOKEN_KEY)
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set("Content-Type", "application/json")
  const token = getToken()
  if (token) headers.set("Authorization", `Bearer ${token}`)
  const res = await fetch(path, { ...init, headers })
  if (res.status === 401) {
    setToken(null)
    if (!path.startsWith("/api/auth/")) window.location.href = "/login"
  }
  if (!res.ok) {
    let msg = res.statusText
    try {
      const body = await res.json()
      msg = body.error || msg
    } catch {
      /* ignore */
    }
    throw new Error(msg)
  }
  if (res.status === 204) return undefined as T
  const ct = res.headers.get("content-type") || ""
  if (ct.includes("text/markdown")) return (await res.text()) as T
  return res.json()
}

export const api = {
  register: (email: string, password: string) =>
    request<{ token: string }>("/api/auth/register", { method: "POST", body: JSON.stringify({ email, password }) }),
  login: (email: string, password: string) =>
    request<{ token: string }>("/api/auth/login", { method: "POST", body: JSON.stringify({ email, password }) }),
  me: () => request<{ id: string; email: string }>("/api/auth/me"),
  projects: () => request<Project[]>("/api/projects"),
  createProject: (body: unknown) => request<Project>("/api/projects", { method: "POST", body: JSON.stringify(body) }),
  getProject: (id: string) => request<Project>(`/api/projects/${id}`),
  updateProject: (id: string, body: unknown) => request<Project>(`/api/projects/${id}`, { method: "PUT", body: JSON.stringify(body) }),
  createBrief: (id: string, prompt: string) =>
    request(`/api/projects/${id}/brief`, { method: "POST", body: JSON.stringify({ prompt }) }),
  analyze: (id: string, prompt?: string) =>
    request<Job>(`/api/projects/${id}/analyze`, { method: "POST", body: JSON.stringify({ prompt }) }),
  job: (projectId: string, jobId: string) => request<Job>(`/api/projects/${projectId}/jobs/${jobId}`),
  latestJob: (projectId: string) => request<Job>(`/api/projects/${projectId}/job`),
  strategy: (id: string) => request<StrategyDoc>(`/api/projects/${id}/strategy`),
  exportUrl: (id: string, format: "json" | "md") => `/api/projects/${id}/export?format=${format}`,
}
