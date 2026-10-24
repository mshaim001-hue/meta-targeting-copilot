package domain

import (
	"time"

	"github.com/google/uuid"
)

type IntentLevel string

const (
	IntentLow    IntentLevel = "low"
	IntentMedium IntentLevel = "medium"
	IntentHigh   IntentLevel = "high"
)

type EntityStatus string

const (
	StatusVerified    EntityStatus = "verified"
	StatusSuggested   EntityStatus = "suggested"
	StatusUnavailable EntityStatus = "unavailable"
	StatusRestricted  EntityStatus = "restricted"
)

type Source string

const (
	SourceLLMHypothesis  Source = "llm_hypothesis"
	SourceMetaVerified   Source = "meta_verified"
	SourceMetaSuggestion Source = "meta_suggestion"
	SourceNotAvailable   Source = "not_available"
)

type StrategyType string

const (
	StrategyBroad        StrategyType = "broad"
	StrategyInterest     StrategyType = "interest"
	StrategyRetargeting  StrategyType = "retargeting"
	StrategyLookalike    StrategyType = "lookalike"
	StrategyExperimental StrategyType = "experimental"
)

type ConfidenceLevel string

const (
	ConfidenceHigh         ConfidenceLevel = "high"
	ConfidenceMedium       ConfidenceLevel = "medium"
	ConfidenceExperimental ConfidenceLevel = "experimental"
)

type JobStage string

const (
	StageQueued                 JobStage = "QUEUED"
	StageAnalyzingBrief         JobStage = "ANALYZING_BRIEF"
	StageCreatingSegments       JobStage = "CREATING_SEGMENTS"
	StageGeneratingSignals      JobStage = "GENERATING_SIGNALS"
	StageVerifyingMetaTargeting JobStage = "VERIFYING_META_TARGETING"
	StageBuildingStrategies     JobStage = "BUILDING_STRATEGIES"
	StageBuildingCreatives      JobStage = "BUILDING_CREATIVES"
	StageComplete               JobStage = "COMPLETE"
)

type FirstPartyAssets struct {
	CustomerList    bool `json:"customerList"`
	Pixel           bool `json:"pixel"`
	CAPI            bool `json:"capi"`
	WebsiteVisitors bool `json:"websiteVisitors"`
	Instagram       bool `json:"instagram"`
	FacebookPage    bool `json:"facebookPage"`
	VideoViews      bool `json:"videoViews"`
	LeadForms       bool `json:"leadForms"`
	Purchasers      bool `json:"purchasers"`
	CRM             bool `json:"crm"`
}

type Project struct {
	ID           uuid.UUID        `json:"id"`
	UserID       uuid.UUID        `json:"userId"`
	BusinessName string           `json:"businessName"`
	Website      string           `json:"website"`
	Product      string           `json:"product"`
	Country      string           `json:"country"`
	City         string           `json:"city"`
	Price        string           `json:"price"`
	Objective    string           `json:"objective"`
	Budget       string           `json:"budget"`
	Assets       FirstPartyAssets `json:"assets"`
	CreatedAt    time.Time        `json:"createdAt"`
}

type BriefInput struct {
	Product        string           `json:"product"`
	AudiencePrompt string           `json:"audiencePrompt"`
	Location       string           `json:"location"`
	Country        string           `json:"country"`
	City           string           `json:"city"`
	Website        string           `json:"website"`
	Price          string           `json:"price"`
	Objective      string           `json:"objective"`
	Budget         string           `json:"budget"`
	Assets         FirstPartyAssets `json:"assets"`
}

type BriefAnalysis struct {
	BusinessModel    string   `json:"businessModel"`
	ConversionGoal   string   `json:"conversionGoal"`
	DesiredAudience  string   `json:"desiredAudience"`
	Intent           string   `json:"intent"`
	Constraints      []string `json:"constraints"`
	MissingInfo      []string `json:"missingInfo"`
	LocalVsOnline    string   `json:"localVsOnline"`
	ClarifyingNeeded bool     `json:"clarifyingNeeded"`
	Questions        []string `json:"questions"`
}

type AudienceSegment struct {
	ID              uuid.UUID   `json:"id,omitempty"`
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	IntentLevel     IntentLevel `json:"intentLevel"`
	Pains           []string    `json:"pains"`
	Motivations     []string    `json:"motivations"`
	SemanticSignals []string    `json:"semanticSignals"`
	OfferAngles     []string    `json:"offerAngles"`
	CreativeAngles  []string    `json:"creativeAngles"`
}

type TargetingEntity struct {
	ID            uuid.UUID    `json:"id,omitempty"`
	Query         string       `json:"query"`
	MetaID        string       `json:"metaId,omitempty"`
	MetaName      string       `json:"metaName,omitempty"`
	MetaType      string       `json:"metaType,omitempty"`
	Status        EntityStatus `json:"status"`
	Source        Source       `json:"source"`
	GeneratedFrom string       `json:"generatedFrom,omitempty"`
	Explanation   string       `json:"explanation,omitempty"`
	AudienceSize  *int64       `json:"audienceSize,omitempty"`
	SegmentName   string       `json:"segmentName,omitempty"`
}

type GeoRecommendation struct {
	Country      string  `json:"country"`
	CountryKey   string  `json:"countryKey,omitempty"`
	City         string  `json:"city"`
	CityKey      string  `json:"cityKey,omitempty"`
	Region       string  `json:"region,omitempty"`
	RadiusKM     int     `json:"radiusKm"`
	AltRadiusKM  int     `json:"altRadiusKm,omitempty"`
	AgeMin       int     `json:"ageMin"`
	AgeMax       int     `json:"ageMax"`
	Gender       string  `json:"gender,omitempty"`
	Language     string  `json:"language,omitempty"`
	Reason       string  `json:"reason"`
	Verified     bool    `json:"verified"`
	LocationType string  `json:"locationType,omitempty"`
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
}

type ReachEstimate struct {
	Available  bool   `json:"available"`
	LowerBound *int64 `json:"lowerBound,omitempty"`
	UpperBound *int64 `json:"upperBound,omitempty"`
	Message    string `json:"message,omitempty"`
}

type QualityScore struct {
	SemanticRelevance int `json:"semanticRelevance"`
	MetaAvailability  int `json:"metaAvailability"`
	AudienceSize      int `json:"audienceSize"`
	IntentStrength    int `json:"intentStrength"`
	FirstPartySignal  int `json:"firstPartySignal"`
	Total             int `json:"total"`
}

type AudienceStrategy struct {
	ID                uuid.UUID         `json:"id,omitempty"`
	Name              string            `json:"name"`
	Type              StrategyType      `json:"type"`
	Geo               GeoRecommendation `json:"geo"`
	Ages              []int             `json:"ages,omitempty"`
	Gender            string            `json:"gender,omitempty"`
	IncludedTargeting []TargetingEntity `json:"includedTargeting"`
	ExcludedTargeting []TargetingEntity `json:"excludedTargeting"`
	Reach             ReachEstimate     `json:"reach"`
	Confidence        float64           `json:"confidence"`
	ConfidenceLevel   ConfidenceLevel   `json:"confidenceLevel"`
	Basis             []string          `json:"basis"`
	Explanation       string            `json:"explanation"`
	Quality           QualityScore      `json:"quality"`
}

type Offer struct {
	SegmentName    string `json:"segmentName"`
	AwarenessLevel string `json:"awarenessLevel"`
	Problem        string `json:"problem"`
	Motivation     string `json:"motivation"`
	Objection      string `json:"objection"`
	Offer          string `json:"offer"`
	Hook           string `json:"hook"`
	CTA            string `json:"cta"`
}

type Creative struct {
	SegmentName string `json:"segmentName"`
	Angle       string `json:"angle"`
	Hook        string `json:"hook"`
	VisualIdea  string `json:"visualIdea"`
	CopyAngle   string `json:"copyAngle"`
	CTA         string `json:"cta"`
	Format      string `json:"format"`
}

type Experiment struct {
	Hypothesis string   `json:"hypothesis"`
	Audience   string   `json:"audience"`
	Variable   string   `json:"variable"`
	Variants   []string `json:"variants"`
}

type CriticFinding struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type PipelineResult struct {
	Brief       BriefInput         `json:"brief"`
	Analysis    BriefAnalysis      `json:"analysis"`
	Segments    []AudienceSegment  `json:"segments"`
	Signals     []string           `json:"semanticSignals"`
	Entities    []TargetingEntity  `json:"verifiedMetaInterests"`
	Geo         GeoRecommendation  `json:"geo"`
	Strategies  []AudienceStrategy `json:"audienceStrategies"`
	Offers      []Offer            `json:"offers"`
	Creatives   []Creative         `json:"creativeAngles"`
	Experiments []Experiment       `json:"experiments"`
	Critic      []CriticFinding    `json:"critic"`
	MetaMode    string             `json:"metaMode"`
	LLMMode     string             `json:"llmMode"`
}
