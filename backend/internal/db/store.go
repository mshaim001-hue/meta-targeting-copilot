package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"meta-targeting-copilot/backend/internal/domain"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	Pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{Pool: pool} }

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (s *Store) CreateUser(ctx context.Context, email, hash string) (User, error) {
	var u User
	err := s.Pool.QueryRow(ctx, `INSERT INTO users (email, password_hash) VALUES ($1,$2) RETURNING id, email, password_hash, created_at`, email, hash).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := s.Pool.QueryRow(ctx, `SELECT id, email, password_hash, created_at FROM users WHERE email=$1`, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	var u User
	err := s.Pool.QueryRow(ctx, `SELECT id, email, password_hash, created_at FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (s *Store) CreateProject(ctx context.Context, p domain.Project) (domain.Project, error) {
	row := s.Pool.QueryRow(ctx, `
		INSERT INTO projects (
			user_id, business_name, website, product, country, city, price, objective, budget,
			has_customer_list, has_pixel, has_capi, has_website_visitors, has_instagram,
			has_facebook_page, has_video_views, has_lead_forms, has_purchasers, has_crm
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		RETURNING id, created_at`,
		p.UserID, p.BusinessName, p.Website, p.Product, p.Country, p.City, p.Price, p.Objective, p.Budget,
		p.Assets.CustomerList, p.Assets.Pixel, p.Assets.CAPI, p.Assets.WebsiteVisitors, p.Assets.Instagram,
		p.Assets.FacebookPage, p.Assets.VideoViews, p.Assets.LeadForms, p.Assets.Purchasers, p.Assets.CRM,
	)
	if err := row.Scan(&p.ID, &p.CreatedAt); err != nil {
		return p, err
	}
	return p, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProject(row scanner) (domain.Project, error) {
	var p domain.Project
	err := row.Scan(
		&p.ID, &p.UserID, &p.BusinessName, &p.Website, &p.Product, &p.Country, &p.City, &p.Price, &p.Objective, &p.Budget,
		&p.Assets.CustomerList, &p.Assets.Pixel, &p.Assets.CAPI, &p.Assets.WebsiteVisitors, &p.Assets.Instagram,
		&p.Assets.FacebookPage, &p.Assets.VideoViews, &p.Assets.LeadForms, &p.Assets.Purchasers, &p.Assets.CRM,
		&p.CreatedAt,
	)
	return p, err
}

const projectCols = `id, user_id, business_name, website, product, country, city, price, objective, budget,
	has_customer_list, has_pixel, has_capi, has_website_visitors, has_instagram,
	has_facebook_page, has_video_views, has_lead_forms, has_purchasers, has_crm, created_at`

func (s *Store) GetProject(ctx context.Context, id uuid.UUID) (domain.Project, error) {
	p, err := scanProject(s.Pool.QueryRow(ctx, `SELECT `+projectCols+` FROM projects WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

func (s *Store) ListProjects(ctx context.Context, userID uuid.UUID) ([]domain.Project, error) {
	rows, err := s.Pool.Query(ctx, `SELECT `+projectCols+` FROM projects WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) UpdateProject(ctx context.Context, p domain.Project) (domain.Project, error) {
	_, err := s.Pool.Exec(ctx, `
		UPDATE projects SET
			business_name=$2, website=$3, product=$4, country=$5, city=$6, price=$7, objective=$8, budget=$9,
			has_customer_list=$10, has_pixel=$11, has_capi=$12, has_website_visitors=$13, has_instagram=$14,
			has_facebook_page=$15, has_video_views=$16, has_lead_forms=$17, has_purchasers=$18, has_crm=$19,
			updated_at=now()
		WHERE id=$1`,
		p.ID, p.BusinessName, p.Website, p.Product, p.Country, p.City, p.Price, p.Objective, p.Budget,
		p.Assets.CustomerList, p.Assets.Pixel, p.Assets.CAPI, p.Assets.WebsiteVisitors, p.Assets.Instagram,
		p.Assets.FacebookPage, p.Assets.VideoViews, p.Assets.LeadForms, p.Assets.Purchasers, p.Assets.CRM,
	)
	return p, err
}

type Brief struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"projectId"`
	Prompt    string    `json:"prompt"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *Store) CreateBrief(ctx context.Context, projectID uuid.UUID, prompt string) (Brief, error) {
	var b Brief
	err := s.Pool.QueryRow(ctx, `INSERT INTO audience_briefs (project_id, prompt) VALUES ($1,$2) RETURNING id, project_id, prompt, created_at`, projectID, prompt).
		Scan(&b.ID, &b.ProjectID, &b.Prompt, &b.CreatedAt)
	return b, err
}

func (s *Store) GetBrief(ctx context.Context, id uuid.UUID) (Brief, error) {
	var b Brief
	err := s.Pool.QueryRow(ctx, `SELECT id, project_id, prompt, created_at FROM audience_briefs WHERE id=$1`, id).
		Scan(&b.ID, &b.ProjectID, &b.Prompt, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return b, ErrNotFound
	}
	return b, err
}

func (s *Store) LatestBrief(ctx context.Context, projectID uuid.UUID) (Brief, error) {
	var b Brief
	err := s.Pool.QueryRow(ctx, `SELECT id, project_id, prompt, created_at FROM audience_briefs WHERE project_id=$1 ORDER BY created_at DESC LIMIT 1`, projectID).
		Scan(&b.ID, &b.ProjectID, &b.Prompt, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return b, ErrNotFound
	}
	return b, err
}

func (s *Store) SaveAnalysis(ctx context.Context, briefID uuid.UUID, analysis domain.BriefAnalysis) error {
	raw, _ := json.Marshal(analysis)
	missing, _ := json.Marshal(analysis.MissingInfo)
	_, err := s.Pool.Exec(ctx, `UPDATE audience_briefs SET analyzed_json=$2, missing_info=$3 WHERE id=$1`, briefID, json.RawMessage(raw), json.RawMessage(missing))
	return err
}

func (s *Store) PersistResult(ctx context.Context, project domain.Project, briefID uuid.UUID, res domain.PipelineResult) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM audience_strategies WHERE brief_id=$1`, briefID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM targeting_entities WHERE brief_id=$1`, briefID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM experiments WHERE brief_id=$1`, briefID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM offers WHERE segment_id IN (SELECT id FROM audience_segments WHERE brief_id=$1)`, briefID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM creatives WHERE segment_id IN (SELECT id FROM audience_segments WHERE brief_id=$1)`, briefID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM audience_segments WHERE brief_id=$1`, briefID); err != nil {
		return err
	}

	segIDs := map[string]uuid.UUID{}
	for i, seg := range res.Segments {
		pains, _ := json.Marshal(seg.Pains)
		mots, _ := json.Marshal(seg.Motivations)
		sigs, _ := json.Marshal(seg.SemanticSignals)
		oa, _ := json.Marshal(seg.OfferAngles)
		ca, _ := json.Marshal(seg.CreativeAngles)
		var id uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO audience_segments (brief_id, name, description, intent_level, pains, motivations, semantic_signals, offer_angles, creative_angles, position)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
			briefID, seg.Name, seg.Description, seg.IntentLevel, json.RawMessage(pains), json.RawMessage(mots), json.RawMessage(sigs), json.RawMessage(oa), json.RawMessage(ca), i,
		).Scan(&id); err != nil {
			return err
		}
		segIDs[seg.Name] = id
	}

	for _, e := range res.Entities {
		var segID *uuid.UUID
		if id, ok := segIDs[e.GeneratedFrom]; ok {
			segID = &id
		}
		raw, _ := json.Marshal(e)
		if _, err := tx.Exec(ctx, `
			INSERT INTO targeting_entities (segment_id, brief_id, query, meta_id, meta_name, meta_type, status, source, generated_from, explanation, audience_size, raw)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			segID, briefID, e.Query, nilIfEmpty(e.MetaID), nilIfEmpty(e.MetaName), nilIfEmpty(e.MetaType),
			e.Status, e.Source, e.GeneratedFrom, e.Explanation, e.AudienceSize, json.RawMessage(raw),
		); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO audit_events (project_id, entity_type, value, source, meta_id, generated_from, payload)
			VALUES ($1,'targeting_entity',$2,$3,$4,$5,$6)`,
			project.ID, e.Query, e.Source, nilIfEmpty(e.MetaID), e.GeneratedFrom, json.RawMessage(raw),
		); err != nil {
			return err
		}
	}

	for _, st := range res.Strategies {
		geo, _ := json.Marshal(st.Geo)
		ages, _ := json.Marshal(st.Ages)
		inc, _ := json.Marshal(st.IncludedTargeting)
		exc, _ := json.Marshal(st.ExcludedTargeting)
		reach, _ := json.Marshal(st.Reach)
		basis, _ := json.Marshal(st.Basis)
		qs, _ := json.Marshal(st.Quality)
		if _, err := tx.Exec(ctx, `
			INSERT INTO audience_strategies (project_id, brief_id, name, type, geo, ages, gender, included_targeting, excluded_targeting, reach_estimate, confidence, confidence_level, basis, explanation, quality_score)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
			project.ID, briefID, st.Name, st.Type, json.RawMessage(geo), json.RawMessage(ages), st.Gender, json.RawMessage(inc), json.RawMessage(exc), json.RawMessage(reach), st.Confidence, st.ConfidenceLevel, json.RawMessage(basis), st.Explanation, json.RawMessage(qs),
		); err != nil {
			return err
		}
	}

	for _, o := range res.Offers {
		id, ok := segIDs[o.SegmentName]
		if !ok {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO offers (segment_id, awareness_level, problem, motivation, objection, offer, hook, cta)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			id, o.AwarenessLevel, o.Problem, o.Motivation, o.Objection, o.Offer, o.Hook, o.CTA,
		); err != nil {
			return err
		}
	}
	for _, c := range res.Creatives {
		id, ok := segIDs[c.SegmentName]
		if !ok {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO creatives (segment_id, angle, hook, visual_idea, copy_angle, cta, format)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			id, c.Angle, c.Hook, c.VisualIdea, c.CopyAngle, c.CTA, c.Format,
		); err != nil {
			return err
		}
	}
	for _, exp := range res.Experiments {
		variants, _ := json.Marshal(exp.Variants)
		if _, err := tx.Exec(ctx, `
			INSERT INTO experiments (project_id, brief_id, hypothesis, audience, variable, variants)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			project.ID, briefID, exp.Hypothesis, exp.Audience, exp.Variable, json.RawMessage(variants),
		); err != nil {
			return err
		}
	}

	analysis, _ := json.Marshal(res.Analysis)
	missing, _ := json.Marshal(res.Analysis.MissingInfo)
	if _, err := tx.Exec(ctx, `UPDATE audience_briefs SET analyzed_json=$2, missing_info=$3 WHERE id=$1`, briefID, json.RawMessage(analysis), json.RawMessage(missing)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) LoadResult(ctx context.Context, project domain.Project, brief Brief) (domain.PipelineResult, error) {
	res := domain.PipelineResult{
		Brief: domain.BriefInput{
			Product: project.Product, AudiencePrompt: brief.Prompt, Location: project.City + ", " + project.Country,
			Country: project.Country, City: project.City, Website: project.Website, Price: project.Price,
			Objective: project.Objective, Budget: project.Budget, Assets: project.Assets,
		},
	}
	var analysisJSON []byte
	if err := s.Pool.QueryRow(ctx, `SELECT analyzed_json FROM audience_briefs WHERE id=$1`, brief.ID).Scan(&analysisJSON); err == nil && len(analysisJSON) > 0 {
		_ = json.Unmarshal(analysisJSON, &res.Analysis)
	}

	rows, err := s.Pool.Query(ctx, `
		SELECT id, name, description, intent_level, pains, motivations, semantic_signals, offer_angles, creative_angles
		FROM audience_segments WHERE brief_id=$1 ORDER BY position`, brief.ID)
	if err != nil {
		return res, err
	}
	defer rows.Close()
	segByID := map[uuid.UUID]string{}
	for rows.Next() {
		var seg domain.AudienceSegment
		var id uuid.UUID
		var pains, mots, sigs, oa, ca []byte
		if err := rows.Scan(&id, &seg.Name, &seg.Description, &seg.IntentLevel, &pains, &mots, &sigs, &oa, &ca); err != nil {
			return res, err
		}
		_ = json.Unmarshal(pains, &seg.Pains)
		_ = json.Unmarshal(mots, &seg.Motivations)
		_ = json.Unmarshal(sigs, &seg.SemanticSignals)
		_ = json.Unmarshal(oa, &seg.OfferAngles)
		_ = json.Unmarshal(ca, &seg.CreativeAngles)
		seg.ID = id
		segByID[id] = seg.Name
		res.Segments = append(res.Segments, seg)
		res.Signals = append(res.Signals, seg.SemanticSignals...)
	}

	erows, err := s.Pool.Query(ctx, `
		SELECT query, meta_id, meta_name, meta_type, status, source, generated_from, explanation, audience_size
		FROM targeting_entities WHERE brief_id=$1 ORDER BY created_at`, brief.ID)
	if err != nil {
		return res, err
	}
	defer erows.Close()
	for erows.Next() {
		var e domain.TargetingEntity
		var metaID, metaName, metaType *string
		if err := erows.Scan(&e.Query, &metaID, &metaName, &metaType, &e.Status, &e.Source, &e.GeneratedFrom, &e.Explanation, &e.AudienceSize); err != nil {
			return res, err
		}
		e.MetaID = deref(metaID)
		e.MetaName = deref(metaName)
		e.MetaType = deref(metaType)
		res.Entities = append(res.Entities, e)
	}

	srows, err := s.Pool.Query(ctx, `
		SELECT name, type, geo, ages, gender, included_targeting, excluded_targeting, reach_estimate, confidence, confidence_level, basis, explanation, quality_score
		FROM audience_strategies WHERE brief_id=$1 ORDER BY created_at`, brief.ID)
	if err != nil {
		return res, err
	}
	defer srows.Close()
	for srows.Next() {
		var st domain.AudienceStrategy
		var geo, ages, inc, exc, reach, basis, qs []byte
		if err := srows.Scan(&st.Name, &st.Type, &geo, &ages, &st.Gender, &inc, &exc, &reach, &st.Confidence, &st.ConfidenceLevel, &basis, &st.Explanation, &qs); err != nil {
			return res, err
		}
		_ = json.Unmarshal(geo, &st.Geo)
		_ = json.Unmarshal(ages, &st.Ages)
		_ = json.Unmarshal(inc, &st.IncludedTargeting)
		_ = json.Unmarshal(exc, &st.ExcludedTargeting)
		_ = json.Unmarshal(reach, &st.Reach)
		_ = json.Unmarshal(basis, &st.Basis)
		_ = json.Unmarshal(qs, &st.Quality)
		res.Strategies = append(res.Strategies, st)
	}

	orows, err := s.Pool.Query(ctx, `
		SELECT o.segment_id, o.awareness_level, o.problem, o.motivation, o.objection, o.offer, o.hook, o.cta
		FROM offers o JOIN audience_segments s ON s.id=o.segment_id WHERE s.brief_id=$1 ORDER BY s.position`, brief.ID)
	if err != nil {
		return res, err
	}
	defer orows.Close()
	for orows.Next() {
		var o domain.Offer
		var sid uuid.UUID
		if err := orows.Scan(&sid, &o.AwarenessLevel, &o.Problem, &o.Motivation, &o.Objection, &o.Offer, &o.Hook, &o.CTA); err != nil {
			return res, err
		}
		o.SegmentName = segByID[sid]
		res.Offers = append(res.Offers, o)
	}

	crows, err := s.Pool.Query(ctx, `
		SELECT c.segment_id, c.angle, c.hook, c.visual_idea, c.copy_angle, c.cta, c.format
		FROM creatives c JOIN audience_segments s ON s.id=c.segment_id WHERE s.brief_id=$1 ORDER BY s.position`, brief.ID)
	if err != nil {
		return res, err
	}
	defer crows.Close()
	for crows.Next() {
		var c domain.Creative
		var sid uuid.UUID
		if err := crows.Scan(&sid, &c.Angle, &c.Hook, &c.VisualIdea, &c.CopyAngle, &c.CTA, &c.Format); err != nil {
			return res, err
		}
		c.SegmentName = segByID[sid]
		res.Creatives = append(res.Creatives, c)
	}

	xrows, err := s.Pool.Query(ctx, `SELECT hypothesis, audience, variable, variants FROM experiments WHERE brief_id=$1`, brief.ID)
	if err != nil {
		return res, err
	}
	defer xrows.Close()
	for xrows.Next() {
		var exp domain.Experiment
		var variants []byte
		if err := xrows.Scan(&exp.Hypothesis, &exp.Audience, &exp.Variable, &variants); err != nil {
			return res, err
		}
		_ = json.Unmarshal(variants, &exp.Variants)
		res.Experiments = append(res.Experiments, exp)
	}

	var geoJSON []byte
	if err := s.Pool.QueryRow(ctx, `SELECT geo FROM audience_strategies WHERE brief_id=$1 ORDER BY created_at LIMIT 1`, brief.ID).Scan(&geoJSON); err == nil && len(geoJSON) > 0 {
		_ = json.Unmarshal(geoJSON, &res.Geo)
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return res, err
	}
	return res, nil
}

type Job struct {
	ID        uuid.UUID  `json:"id"`
	ProjectID uuid.UUID  `json:"projectId"`
	BriefID   *uuid.UUID `json:"briefId,omitempty"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	Stage     string     `json:"stage"`
	Error     *string    `json:"error,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (s *Store) EnqueueJob(ctx context.Context, projectID, briefID uuid.UUID, jobType string) (Job, error) {
	var j Job
	err := s.Pool.QueryRow(ctx, `
		INSERT INTO jobs (project_id, brief_id, type, status, stage)
		VALUES ($1,$2,$3,'queued','QUEUED')
		RETURNING id, project_id, brief_id, type, status, stage, error, created_at, updated_at`,
		projectID, briefID, jobType,
	).Scan(&j.ID, &j.ProjectID, &j.BriefID, &j.Type, &j.Status, &j.Stage, &j.Error, &j.CreatedAt, &j.UpdatedAt)
	return j, err
}

func (s *Store) GetJob(ctx context.Context, id uuid.UUID) (Job, error) {
	var j Job
	err := s.Pool.QueryRow(ctx, `SELECT id, project_id, brief_id, type, status, stage, error, created_at, updated_at FROM jobs WHERE id=$1`, id).
		Scan(&j.ID, &j.ProjectID, &j.BriefID, &j.Type, &j.Status, &j.Stage, &j.Error, &j.CreatedAt, &j.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return j, ErrNotFound
	}
	return j, err
}

func (s *Store) LatestJob(ctx context.Context, projectID uuid.UUID) (Job, error) {
	var j Job
	err := s.Pool.QueryRow(ctx, `SELECT id, project_id, brief_id, type, status, stage, error, created_at, updated_at FROM jobs WHERE project_id=$1 ORDER BY created_at DESC LIMIT 1`, projectID).
		Scan(&j.ID, &j.ProjectID, &j.BriefID, &j.Type, &j.Status, &j.Stage, &j.Error, &j.CreatedAt, &j.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return j, ErrNotFound
	}
	return j, err
}

func (s *Store) ClaimJobs(ctx context.Context, limit int) ([]Job, error) {
	rows, err := s.Pool.Query(ctx, `
		UPDATE jobs SET status='running', locked_at=now(), attempts=attempts+1, updated_at=now()
		WHERE id IN (
			SELECT id FROM jobs WHERE status='queued' AND run_after<=now()
			ORDER BY created_at LIMIT $1 FOR UPDATE SKIP LOCKED
		)
		RETURNING id, project_id, brief_id, type, status, stage, error, created_at, updated_at`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.ProjectID, &j.BriefID, &j.Type, &j.Status, &j.Stage, &j.Error, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (s *Store) UpdateJobStage(ctx context.Context, id uuid.UUID, stage, status string, jobErr *string) error {
	_, err := s.Pool.Exec(ctx, `UPDATE jobs SET stage=$2, status=$3, error=$4, updated_at=now() WHERE id=$1`, id, stage, status, jobErr)
	return err
}

func (s *Store) CompleteJob(ctx context.Context, id uuid.UUID, result any) error {
	raw, _ := json.Marshal(result)
	_, err := s.Pool.Exec(ctx, `UPDATE jobs SET status='complete', stage='COMPLETE', result=$2, error=NULL, updated_at=now() WHERE id=$1`, id, json.RawMessage(raw))
	return err
}

func (s *Store) FailJob(ctx context.Context, id uuid.UUID, msg string) error {
	_, err := s.Pool.Exec(ctx, `UPDATE jobs SET status='failed', error=$2, updated_at=now() WHERE id=$1`, id, msg)
	return err
}

func (s *Store) CacheGet(ctx context.Context, query, country, kind, version string) (json.RawMessage, time.Time, bool, error) {
	var raw json.RawMessage
	var ts time.Time
	err := s.Pool.QueryRow(ctx, `SELECT result, created_at FROM targeting_cache WHERE query=$1 AND country=$2 AND kind=$3 AND api_version=$4`, query, country, kind, version).
		Scan(&raw, &ts)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, time.Time{}, false, nil
	}
	if err != nil {
		return nil, time.Time{}, false, err
	}
	return raw, ts, true, nil
}

func (s *Store) CachePut(ctx context.Context, query, country, kind, version string, result json.RawMessage) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO targeting_cache (query, country, kind, result, api_version)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (query, country, kind, api_version)
		DO UPDATE SET result=EXCLUDED.result, created_at=now()`, query, country, kind, result, version)
	return err
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
