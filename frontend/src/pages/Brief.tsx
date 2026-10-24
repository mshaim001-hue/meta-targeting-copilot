import { useEffect, useState, type FormEvent } from "react"
import { useNavigate, useParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Label, Textarea } from "@/components/ui/input"
import { api } from "@/lib/api"
import type { Project } from "@/lib/types"

export function BriefPage() {
  const { id } = useParams()
  const nav = useNavigate()
  const [project, setProject] = useState<Project | null>(null)
  const [prompt, setPrompt] = useState("У меня фитнес-клуб в Астане. Хочу привлечь людей 25–40 лет, которые уже занимаются спортом и думают поменять зал.")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)

  useEffect(() => {
    if (!id) return
    api.getProject(id).then(setProject).catch((e: Error) => setError(e.message))
  }, [id])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!id) return
    setPending(true)
    setError("")
    try {
      const job = await api.analyze(id, prompt)
      nav(`/projects/${id}/analyze/${job.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed")
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="max-w-3xl">
      <p className="text-xs uppercase tracking-[0.2em] text-ink/50">
        {project ? `${project.businessName} · ${project.city}` : "Audience"}
      </p>
      <h1 className="mt-2 font-serif text-4xl">Who do you want to reach?</h1>
      <p className="mt-3 max-w-xl text-ink/70">
        Describe the audience in natural language. The LLM will think in marketing terms. Meta API will decide what is actually targetable.
      </p>
      <form className="mt-8 space-y-4" onSubmit={onSubmit}>
        <Card>
          <Label>Audience prompt</Label>
          <Textarea
            className="mt-2 min-h-40 font-serif text-lg"
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="People in Astana who already exercise regularly and are considering changing gyms."
          />
        </Card>
        {error ? <p className="text-sm text-missing">{error}</p> : null}
        <Button type="submit" size="lg" disabled={pending}>
          {pending ? "Queuing…" : "Build Strategy"}
        </Button>
      </form>
    </div>
  )
}
