import { Badge } from "@/components/ui/badge"
import type { TargetingEntity } from "@/lib/types"

const styles: Record<string, string> = {
  meta_verified: "border-verified/30 bg-verified/10 text-verified",
  meta_suggestion: "border-cobalt/30 bg-cobalt/10 text-cobalt",
  llm_hypothesis: "border-hypothesis/30 bg-hypothesis/10 text-hypothesis",
  not_available: "border-missing/30 bg-missing/10 text-missing",
}

const labels: Record<string, string> = {
  meta_verified: "Meta verified",
  meta_suggestion: "Meta suggestion",
  llm_hypothesis: "LLM hypothesis",
  not_available: "Not available",
}

export function SourceBadge({ source }: { source: TargetingEntity["source"] | string }) {
  return <Badge className={styles[source] || ""}>{labels[source] || source}</Badge>
}

export function EntityChip({ entity }: { entity: TargetingEntity }) {
  const name = entity.metaName || entity.query
  return (
    <div className="flex flex-wrap items-center gap-2 rounded-lg border border-rule bg-paper px-3 py-2">
      <span className="text-sm">
        {entity.status === "verified" ? "✓ " : entity.status === "unavailable" ? "× " : "? "}
        {name}
      </span>
      <SourceBadge source={entity.source} />
      {entity.metaId ? <span className="font-mono text-[10px] text-ink/50">{entity.metaId}</span> : null}
    </div>
  )
}
