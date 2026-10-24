import { Link, Outlet, useNavigate } from "react-router-dom"
import { getToken, setToken } from "@/lib/api"

export function Layout() {
  const nav = useNavigate()
  const authed = Boolean(getToken())
  return (
    <div className="min-h-screen">
      <header className="border-b border-rule">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <Link to={authed ? "/projects" : "/"} className="font-serif text-xl tracking-tight">
            Targeting Copilot
          </Link>
          {authed ? (
            <button
              className="text-sm text-ink/60 hover:text-ink"
              onClick={() => {
                setToken(null)
                nav("/login")
              }}
            >
              Sign out
            </button>
          ) : (
            <Link to="/login" className="text-sm text-ink/60 hover:text-ink">
              Sign in
            </Link>
          )}
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-6 py-10">
        <Outlet />
      </main>
    </div>
  )
}
