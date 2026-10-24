package ai

import (
	"encoding/json"
	"fmt"

	"meta-targeting-copilot/backend/internal/domain"
)

const systemRules = `You are a senior performance marketer specializing in Meta Ads.
Return ONLY valid JSON. No markdown. No commentary.
Never invent Meta interest IDs, Meta location keys, or audience size numbers.
Never claim a targeting option exists in Meta. You produce marketing hypotheses and semantic signals only.
Write user-facing copy in the same language as the audience prompt.`

func briefAnalyzerPrompt(in domain.BriefInput) string {
	b, _ := json.MarshalIndent(in, "", "  ")
	return systemRules + `

Task: Brief Analyzer.
Extract the business model, conversion goal, desired audience, intent, constraints, and missing information.
Ask at most 5 clarifying questions, only if they materially change targeting (first-party data, local vs online, pixel).
JSON schema:
{
  "businessModel": "string",
  "conversionGoal": "string",
  "desiredAudience": "string",
  "intent": "string",
  "constraints": ["string"],
  "missingInfo": ["string"],
  "localVsOnline": "local|online|hybrid",
  "clarifyingNeeded": false,
  "questions": ["string"]
}

Input:
` + string(b)
}

func segmentationPrompt(in domain.BriefInput, analysis domain.BriefAnalysis) string {
	payload, _ := json.MarshalIndent(map[string]any{"brief": in, "analysis": analysis}, "", "  ")
	return systemRules + `

Task: Audience Segmentation Agent.
Create 3 to 7 DISTINCT marketing segments from the desired audience.
Do not use Meta API. Do not invent Meta targeting names as if they were confirmed.
Each segment needs pains, motivations, semanticSignals (searchable phrases), offerAngles, creativeAngles.
intentLevel must be one of: low, medium, high.
JSON schema:
{
  "segments": [
    {
      "name": "string",
      "description": "string",
      "intentLevel": "high",
      "pains": ["string"],
      "motivations": ["string"],
      "semanticSignals": ["string"],
      "offerAngles": ["string"],
      "creativeAngles": ["string"]
    }
  ]
}

Input:
` + string(payload)
}

func signalPrompt(segs []domain.AudienceSegment) string {
	payload, _ := json.MarshalIndent(segs, "", "  ")
	return systemRules + `

Task: Signal Generator.
For each segment, expand semantic signals into short English phrases that could be searched in a targeting catalog
(interests, behaviors, brands, activities). 5-12 signals per segment. No Meta IDs.
JSON schema:
{
  "segments": [
    {"name": "string", "signals": ["string"]}
  ]
}

Input:
` + string(payload)
}

func geoPrompt(in domain.BriefInput, analysis domain.BriefAnalysis) string {
	payload, _ := json.MarshalIndent(map[string]any{"brief": in, "analysis": analysis}, "", "  ")
	return systemRules + `

Task: Geo Strategy Agent.
Recommend country, city, radiusKm, optional altRadiusKm, ageMin, ageMax, gender (or empty), language, and a reason.
Do not invent Meta location keys. Leave countryKey/cityKey empty.
JSON schema:
{
  "country": "string",
  "city": "string",
  "radiusKm": 8,
  "altRadiusKm": 12,
  "ageMin": 23,
  "ageMax": 42,
  "gender": "",
  "language": "",
  "reason": "string"
}

Input:
` + string(payload)
}

func offerPrompt(in domain.BriefInput, segs []domain.AudienceSegment) string {
	payload, _ := json.MarshalIndent(map[string]any{"brief": in, "segments": segs}, "", "  ")
	return systemRules + `

Task: Offer Strategist.
One offer per segment. Do not talk to the audience as if they are beginners if the brief says they already do the activity.
JSON schema:
{
  "offers": [
    {
      "segmentName": "string",
      "awarenessLevel": "string",
      "problem": "string",
      "motivation": "string",
      "objection": "string",
      "offer": "string",
      "hook": "string",
      "cta": "string"
    }
  ]
}

Input:
` + string(payload)
}

func creativePrompt(in domain.BriefInput, segs []domain.AudienceSegment, offers []domain.Offer) string {
	payload, _ := json.MarshalIndent(map[string]any{"brief": in, "segments": segs, "offers": offers}, "", "  ")
	return systemRules + fmt.Sprintf(`

Task: Creative Strategist.
One primary creative per segment. format is one of: reel, story, feed, carousel, static.
Map Audience → Problem → Message → Creative.
JSON schema:
{
  "creatives": [
    {
      "segmentName": "string",
      "angle": "string",
      "hook": "string",
      "visualIdea": "string",
      "copyAngle": "string",
      "cta": "string",
      "format": "reel"
    }
  ]
}

Input:
%s`, payload)
}

func criticPrompt(payload string) string {
	return systemRules + `

Task: Strategy Critic.
Find hallucinated targeting, overlapping segments, weak offers, too many AND conditions, contradictions.
Severity: info | warning | error.
JSON schema:
{"findings":[{"severity":"warning","message":"string"}]}

Strategy:
` + payload
}
