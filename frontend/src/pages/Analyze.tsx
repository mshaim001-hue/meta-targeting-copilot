import { useEffect, useState } from "react"
import { Link, useNavigate, useParams } from "react-router-dom"
import { Card } from "@/components/ui/card"
import { api } from "@/lib/api"
import type { Job } from "@/lib/types"

const stages = [
  "QUEUED",
  "ANALYZING_BRIEF",
  "CREATING_SEGMENTS",
  "GENERATING_SIGNALS",
  "VERIFYING_META_TARGETING",
  "BUILDING_STRATEGIES",
  "BUILDING_CREATIVES",
  "COMPLETE",
]

export function AnalyzePage() {
  const { id, jobId } = useParams()
  const nav = useNavigate()
  const [job, setJob] = useState<Job | null>(null)

  useEffect(() => {
    if (!id || !jobId) return
    let stop = false
    async function tick() {
      try {
        const j = await api.job(id!, jobId!)
        if (stop) return
        setJob(j)
        if (j.status === "complete") {
          nav(`/projects/${id}/strategy`)
          return
        }
        if (j.status === "failed") return
        window.setTimeout(tick, 900)
      } catch {
        window.setTimeout(tick, 1500)
      }
    }
    tick()
    return () => {
      stop = true
    }
  }, [id, jobId, nav])

  const idx = Math.max(0, stages.indexOf(job?.stage || "QUEUED"))

  return (
    <div className="max-w-2xl">
      <p className="text-xs uppercase tracking-[0.2em] text-ink/50">Analysis</p>
      <h1 className="mt-2 font-serif text-4xl">Building a grounded strategy</h1>
      <Card className="mt-8 space-y-3">
        {stages.filter((s) => s !== "QUEUED").map((s, i) => {
          const done = i < idx
          const current = stages[idx] === s
          return (
            <div key={s} className="flex items-center gap-3 text-sm">
              <span className={`h-2 w-2 rounded-full ${done || current ? "bg-cobalt" : "bg-rule"}`} />
              <span className={current ? "font-medium" : "text-ink/60"}>{s.replaceAll("_", " ")}</span>
            </div>
          )
        })}
      </Card>
      {job?.status === "failed" ? (
        <p className="mt-6 text-sm text-missing">
          {job.error} · <Link className="underline" to={`/projects/${id}`}>Back</Link>
        </p>
      ) : (
        <p className="mt-6 text-sm text-ink/50">LLM hypotheses are not marked as Meta targeting until the resolver confirms them.</p>
      )}
    </div>
  )
}
