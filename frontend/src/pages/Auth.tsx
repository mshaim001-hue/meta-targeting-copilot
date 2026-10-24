import { useState, type FormEvent } from "react"
import { Link, useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Input, Label } from "@/components/ui/input"
import { api, setToken } from "@/lib/api"

export function AuthPage({ mode }: { mode: "login" | "register" }) {
  const nav = useNavigate()
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setPending(true)
    setError("")
    try {
      const res = mode === "login" ? await api.login(email, password) : await api.register(email, password)
      setToken(res.token)
      nav("/projects")
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed")
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="mx-auto max-w-md">
      <h1 className="font-serif text-4xl">{mode === "login" ? "Sign in" : "Create account"}</h1>
      <p className="mt-2 text-sm text-ink/60">JWT auth lives on the Go API. No third-party identity provider required.</p>
      <Card className="mt-8">
        <form className="space-y-4" onSubmit={onSubmit}>
          <div className="space-y-1">
            <Label>Email</Label>
            <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
          </div>
          <div className="space-y-1">
            <Label>Password</Label>
            <Input type="password" minLength={8} value={password} onChange={(e) => setPassword(e.target.value)} required />
          </div>
          {error ? <p className="text-sm text-missing">{error}</p> : null}
          <Button type="submit" disabled={pending} className="w-full">
            {pending ? "Please wait…" : mode === "login" ? "Sign in" : "Register"}
          </Button>
        </form>
      </Card>
      <p className="mt-4 text-sm text-ink/60">
        {mode === "login" ? (
          <>No account? <Link to="/register" className="underline">Register</Link></>
        ) : (
          <>Already registered? <Link to="/login" className="underline">Sign in</Link></>
        )}
      </p>
    </div>
  )
}
