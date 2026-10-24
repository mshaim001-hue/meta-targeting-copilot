export type Assets = {
  customerList: boolean
  pixel: boolean
  capi: boolean
  websiteVisitors: boolean
  instagram: boolean
  facebookPage: boolean
  videoViews: boolean
  leadForms: boolean
  purchasers: boolean
  crm: boolean
}

export type Project = {
  id: string
  businessName: string
  website: string
  product: string
  country: string
  city: string
  price: string
  objective: string
  budget: string
  assets: Assets
  createdAt: string
}

export type Job = {
  id: string
  projectId: string
  briefId?: string
  type: string
  status: string
  stage: string
  error?: string | null
}

export type TargetingEntity = {
  query: string
  metaId?: string
  metaName?: string
  metaType?: string
  status: "verified" | "suggested" | "unavailable" | "restricted"
  source: "llm_hypothesis" | "meta_verified" | "meta_suggestion" | "not_available"
  generatedFrom?: string
  explanation?: string
  audienceSize?: number
}

export type Segment = {
  id?: string
  name: string
  description: string
  intentLevel: "low" | "medium" | "high"
  pains: string[]
  motivations: string[]
  semanticSignals: string[]
  offerAngles: string[]
  creativeAngles: string[]
}

export type StrategyDoc = {
  brief: { product: string; audiencePrompt: string; country: string; city: string }
  analysis: { businessModel: string; conversionGoal: string; desiredAudience: string; questions?: string[]; missingInfo?: string[] }
  segments: Segment[]
  verifiedMetaInterests: TargetingEntity[]
  geo: {
    country: string
    city: string
    countryKey?: string
    cityKey?: string
    radiusKm: number
    altRadiusKm?: number
    ageMin: number
    ageMax: number
    reason: string
    verified: boolean
  }
  audienceStrategies: Array<{
    name: string
    type: string
    geo: StrategyDoc["geo"]
    includedTargeting: TargetingEntity[]
    excludedTargeting: TargetingEntity[]
    reach: { available: boolean; lowerBound?: number; upperBound?: number; message?: string }
    confidence: number
    confidenceLevel: string
    basis: string[]
    explanation: string
    quality: { semanticRelevance: number; metaAvailability: number; audienceSize: number; intentStrength: number; firstPartySignal: number; total: number }
  }>
  offers: Array<{ segmentName: string; awarenessLevel: string; problem: string; motivation: string; objection: string; offer: string; hook: string; cta: string }>
  creativeAngles: Array<{ segmentName: string; angle: string; hook: string; visualIdea: string; copyAngle: string; cta: string; format: string }>
  experiments: Array<{ hypothesis: string; audience: string; variable: string; variants: string[] }>
  critic: Array<{ severity: string; message: string }>
  metaMode: string
  llmMode: string
}

export const emptyAssets = (): Assets => ({
  customerList: false,
  pixel: false,
  capi: false,
  websiteVisitors: false,
  instagram: false,
  facebookPage: false,
  videoViews: false,
  leadForms: false,
  purchasers: false,
  crm: false,
})
