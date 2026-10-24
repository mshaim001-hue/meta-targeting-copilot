import { useEffect, useState } from "react"
import { Link, useParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { EntityChip, SourceBadge } from "@/components/SourceBadge"
import { Badge } from "@/components/ui/badge"
import { getToken } from "@/lib/api"
import { api } from "@/lib/api"
import type { StrategyDoc } from "@/lib/types"
import { formatReach } from "@/lib/utils"

const tabs = [
  "Overview",
  "Segments",
  "Offer",
  "Geo & Demographics",
  "Meta Targeting",
  "Retargeting & LAL",
  "Creative Strategy",
  "Experiments",
] as const

export function StrategyPage() {
  const { id } = useParams()
  const [doc, setDoc] = useState<StrategyDoc | null>(null)
  const [tab, setTab] = useState<(typeof tabs)[number]>("Overview")
  const [error, setError] = useState("")

  useEffect(() => {
    if (!id) return
    api.strategy(id).then(setDoc).catch((e: Error) => setError(e.message))
  }, [id])

  async function download(format: "json" | "md") {
    if (!id) return
    const res = await fetch(`/api/projects/${id}/export?format=${format}`, {
      headers: { Authorization: `Bearer ${getToken()}` },
    })
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = format === "md" ? "strategy.md" : "strategy.json"
    a.click()
    URL.revokeObjectURL(url)
  }

  if (error) {
    return (
      <div>
        <p className="text-missing">{error}</p>
        <Link className="mt-4 inline-block underline" to={`/projects/${id}`}>Write a brief first</Link>
      </div>
    )
  }
  if (!doc) return <p className="text-ink/50">Loading strategy…</p>

  return (
    <div>
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-ink/50">{doc.brief.city} · {doc.brief.product}</p>
          <h1 className="mt-2 font-serif text-4xl">Strategy</h1>
          <p className="mt-2 max-w-2xl text-sm text-ink/60">{doc.brief.audiencePrompt}</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => download("md")}>Export Markdown</Button>
          <Button variant="outline" onClick={() => download("json")}>Export JSON</Button>
          <Button asChild variant="ghost"><Link to={`/projects/${id}`}>New brief</Link></Button>
        </div>
      </div>

      <div className="mt-4 flex flex-wrap gap-2 text-xs text-ink/50">
        <span>LLM: {doc.llmMode}</span>
        <span>Meta: {doc.metaMode}</span>
        {!doc.metaMode.startsWith("meta-graph") ? (
          <Badge className="border-hypothesis/30 text-hypothesis">Mock Meta catalog — connect an access token to verify live</Badge>
        ) : null}
      </div>

      <div className="mt-8 flex flex-wrap gap-2 border-b border-rule pb-3">
        {tabs.map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`rounded-full px-3 py-1 text-sm ${tab === t ? "bg-ink text-paper" : "text-ink/60 hover:bg-paper-2"}`}
          >
            {t}
          </button>
        ))}
      </div>

      <div className="mt-8">
        {tab === "Overview" && <Overview doc={doc} />}
        {tab === "Segments" && <Segments doc={doc} />}
        {tab === "Offer" && <Offers doc={doc} />}
        {tab === "Geo & Demographics" && <Geo doc={doc} />}
        {tab === "Meta Targeting" && <Meta doc={doc} />}
        {tab === "Retargeting & LAL" && <Retargeting doc={doc} />}
        {tab === "Creative Strategy" && <Creatives doc={doc} />}
        {tab === "Experiments" && <Experiments doc={doc} />}
      </div>
    </div>
  )
}

function Overview({ doc }: { doc: StrategyDoc }) {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      {doc.audienceStrategies.map((s) => (
        <Card key={s.name}>
          <div className="flex items-center justify-between">
            <h2 className="font-serif text-2xl">{s.name}</h2>
            <Badge>{s.confidenceLevel}</Badge>
          </div>
          <p className="mt-2 text-sm text-ink/70">{s.explanation}</p>
          <p className="mt-4 text-sm">
            Reach: {s.reach.available ? formatReach(s.reach.lowerBound, s.reach.upperBound) : "Estimate unavailable"}
          </p>
          <p className="mt-1 text-xs text-ink/50">Quality score {s.quality.total}/100 — product heuristic, not a Meta metric</p>
        </Card>
      ))}
      {doc.critic?.length ? (
        <Card className="md:col-span-2">
          <h3 className="font-medium">Critic</h3>
          <ul className="mt-2 space-y-1 text-sm">
            {doc.critic.map((c, i) => (
              <li key={i}><span className="uppercase text-xs text-ink/50">{c.severity}</span> — {c.message}</li>
            ))}
          </ul>
        </Card>
      ) : null}
    </div>
  )
}

function Segments({ doc }: { doc: StrategyDoc }) {
  return (
    <div className="grid gap-4">
      {doc.segments.map((seg) => {
        const ents = doc.verifiedMetaInterests.filter((e) => e.generatedFrom === seg.name)
        const offer = doc.offers.find((o) => o.segmentName === seg.name)
        return (
          <Card key={seg.name}>
            <div className="flex flex-wrap items-start justify-between gap-2">
              <h2 className="font-serif text-2xl">{seg.name}</h2>
              <Badge>Intent {seg.intentLevel}</Badge>
            </div>
            <p className="mt-2 text-sm text-ink/70">{seg.description}</p>
            <p className="mt-4 text-xs uppercase tracking-widest text-ink/50">Meta signals</p>
            <div className="mt-2 flex flex-wrap gap-2">
              {ents.length ? ents.map((e) => <EntityChip key={e.query} entity={e} />) : <span className="text-sm text-ink/50">No Meta match yet</span>}
            </div>
            <p className="mt-4 text-sm"><span className="text-ink/50">Geo</span> {doc.geo.city} · {doc.geo.ageMin}–{doc.geo.ageMax}</p>
            {offer ? <p className="mt-2 text-sm"><span className="text-ink/50">Offer</span> {offer.offer}</p> : null}
          </Card>
        )
      })}
    </div>
  )
}

function Offers({ doc }: { doc: StrategyDoc }) {
  return (
    <div className="grid gap-4">
      {doc.offers.map((o) => (
        <Card key={o.segmentName}>
          <h2 className="font-serif text-2xl">{o.segmentName}</h2>
          <dl className="mt-3 grid gap-2 text-sm md:grid-cols-2">
            <Row k="Problem" v={o.problem} />
            <Row k="Motivation" v={o.motivation} />
            <Row k="Objection" v={o.objection} />
            <Row k="Offer" v={o.offer} />
            <Row k="Hook" v={o.hook} />
            <Row k="CTA" v={o.cta} />
          </dl>
        </Card>
      ))}
    </div>
  )
}

function Geo({ doc }: { doc: StrategyDoc }) {
  const g = doc.geo
  return (
    <Card>
      <h2 className="font-serif text-2xl">{g.city}, {g.country}</h2>
      <p className="mt-2 text-sm">{g.reason}</p>
      <div className="mt-4 flex flex-wrap gap-2">
        <SourceBadge source={g.verified ? "meta_verified" : "llm_hypothesis"} />
        {g.cityKey ? <Badge className="font-mono">{g.cityKey}</Badge> : null}
        {g.countryKey ? <Badge className="font-mono">{g.countryKey}</Badge> : null}
      </div>
      <p className="mt-4 text-sm">Radius {g.radiusKm} km {g.altRadiusKm ? `· test ${g.altRadiusKm} km` : ""} · Age {g.ageMin}–{g.ageMax}</p>
    </Card>
  )
}

function Meta({ doc }: { doc: StrategyDoc }) {
  return (
    <div className="space-y-3">
      <p className="text-sm text-ink/60">
        LLM never marks an interest as available. Each row is a marketing signal plus the Meta resolver verdict.
      </p>
      {doc.verifiedMetaInterests.map((e) => (
        <Card key={e.query + e.generatedFrom}>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div>
              <p className="font-medium">{e.query}</p>
              {e.metaName && e.metaName !== e.query ? <p className="text-sm text-ink/60">Meta: {e.metaName}</p> : null}
            </div>
            <SourceBadge source={e.source} />
          </div>
          <p className="mt-2 text-sm text-ink/70">{e.explanation}</p>
          <p className="mt-1 text-xs text-ink/50">Why? Generated from segment “{e.generatedFrom}”.</p>
        </Card>
      ))}
    </div>
  )
}

function Retargeting({ doc }: { doc: StrategyDoc }) {
  const items = doc.audienceStrategies.filter((s) => s.type === "retargeting" || s.type === "lookalike")
  return (
    <div className="grid gap-4">
      {items.map((s) => (
        <Card key={s.name}>
          <h2 className="font-serif text-2xl">{s.name}</h2>
          <p className="mt-2 text-sm text-ink/70">{s.explanation}</p>
          <div className="mt-3 space-y-2">
            {s.includedTargeting.map((e) => <EntityChip key={e.query} entity={e} />)}
            {s.excludedTargeting.map((e) => <EntityChip key={`x-${e.query}`} entity={e} />)}
          </div>
        </Card>
      ))}
    </div>
  )
}

function Creatives({ doc }: { doc: StrategyDoc }) {
  return (
    <div className="grid gap-4">
      {doc.creativeAngles.map((c) => {
        const offer = doc.offers.find((o) => o.segmentName === c.segmentName)
        return (
          <Card key={c.segmentName}>
            <p className="text-xs uppercase tracking-widest text-ink/50">Audience → Problem → Message → Creative</p>
            <h2 className="mt-2 font-serif text-2xl">{c.segmentName}</h2>
            <p className="mt-2 text-sm">Problem: {offer?.problem}</p>
            <p className="text-sm">Message: {c.copyAngle}</p>
            <p className="mt-3 text-sm"><span className="text-ink/50">Angle</span> {c.angle} · {c.format}</p>
            <p className="mt-2 font-serif text-xl">“{c.hook}”</p>
            <p className="mt-2 text-sm text-ink/70">{c.visualIdea}</p>
          </Card>
        )
      })}
    </div>
  )
}

function Experiments({ doc }: { doc: StrategyDoc }) {
  return (
    <div className="grid gap-4">
      {doc.experiments.map((e, i) => (
        <Card key={i}>
          <h2 className="font-serif text-2xl">Experiment {String(i + 1).padStart(3, "0")}</h2>
          <p className="mt-2 text-sm">{e.hypothesis}</p>
          <p className="mt-3 text-sm">Audience: {e.audience}</p>
          <p className="text-sm">Variable: {e.variable} — do not change more than one thing at a time.</p>
          <ul className="mt-2 list-disc pl-5 text-sm">
            {e.variants.map((v) => <li key={v}>{v}</li>)}
          </ul>
        </Card>
      ))}
    </div>
  )
}

function Row({ k, v }: { k: string; v: string }) {
  return (
    <div>
      <dt className="text-xs uppercase tracking-widest text-ink/50">{k}</dt>
      <dd className="mt-1">{v}</dd>
    </div>
  )
}
