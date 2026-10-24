import { useEffect, useState } from "react"
import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { api } from "@/lib/api"
import type { Project } from "@/lib/types"

export function ProjectsPage() {
  const [items, setItems] = useState<Project[]>([])
  const [error, setError] = useState("")

  useEffect(() => {
    api.projects().then(setItems).catch((e: Error) => setError(e.message))
  }, [])

  return (
    <div>
      <div className="flex items-end justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-ink/50">Projects</p>
          <h1 className="mt-2 font-serif text-4xl">Who do you want to find?</h1>
        </div>
        <Button asChild>
          <Link to="/projects/new">New project</Link>
        </Button>
      </div>
      {error ? <p className="mt-6 text-sm text-missing">{error}</p> : null}
      <div className="mt-8 grid gap-4 md:grid-cols-2">
        {items.map((p) => (
          <Link key={p.id} to={`/projects/${p.id}`}>
            <Card className="h-full hover:border-ink/30">
              <p className="text-xs uppercase tracking-widest text-ink/50">{p.city}, {p.country}</p>
              <h2 className="mt-2 font-serif text-2xl">{p.businessName}</h2>
              <p className="mt-1 text-sm text-ink/70">{p.product}</p>
            </Card>
          </Link>
        ))}
        {items.length === 0 && !error ? (
          <Card>
            <p className="text-sm text-ink/60">No projects yet. Create one and describe the audience in natural language.</p>
          </Card>
        ) : null}
      </div>
    </div>
  )
}
