package model

import "time"

// Users and sessions have PostgreSQL migrations; other records remain storage designs.
// IDs are application-assigned strings; PostgreSQL column types will be set by migrations.
// Nested values are intended for JSONB and need explicit encoding in a future adapter.
// Records must be mapped to response DTOs, never returned directly from handlers.
type UserRecord struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	DisplayName  string    `db:"display_name"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// One current profile per user. Version supports recommendation traceability.
type ProfileRecord struct {
	UserID    string    `db:"user_id"`
	Version   int64     `db:"version"`
	Profile   Profile   `db:"profile"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Evidence retains the source; task completion alone is not verified skill growth.
type SkillEvidenceRecord struct {
	ID         string     `db:"id"`
	UserID     string     `db:"user_id"`
	SkillID    string     `db:"skill_id"`
	SourceType string     `db:"source_type"` // self-assessment, assessment, task-result
	SourceID   *string    `db:"source_id"`
	Level      float64    `db:"level"`
	Confidence float64    `db:"confidence"`
	ObservedAt time.Time  `db:"observed_at"`
	VerifiedAt *time.Time `db:"verified_at"`
}

type TaskRecord struct {
	ID               string    `db:"id"`
	Version          int64     `db:"version"`
	Task             Task      `db:"task"`
	PrerequisiteIDs  []string  `db:"prerequisite_ids"`
	SourceURL        *string   `db:"source_url"`
	SourceConfidence float64   `db:"source_confidence"`
	Available        bool      `db:"available"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

// Snapshots preserve scoring inputs even after profile/task/config changes.
type RecommendationRecord struct {
	ID              string             `db:"id"`
	UserID          string             `db:"user_id"`
	TaskID          string             `db:"task_id"`
	ProfileVersion  int64              `db:"profile_version"`
	TaskVersion     int64              `db:"task_version"`
	ProfileSnapshot Profile            `db:"profile_snapshot"`
	Result          Recommendation     `db:"result"`
	RuleConfig      map[string]float64 `db:"rule_config"`
	CreatedAt       time.Time          `db:"created_at"`
}

// Future uniqueness is (user_id, event_id), not a global client-supplied event ID.
type FeedbackRecord struct {
	UserID     string    `db:"user_id"`
	EventID    string    `db:"event_id"`
	TargetType string    `db:"target_type"`
	TargetID   string    `db:"target_id"`
	EventType  string    `db:"event_type"`
	Value      *float64  `db:"value"`
	OccurredAt time.Time `db:"occurred_at"`
	ReceivedAt time.Time `db:"received_at"`
}

// Domain extensions reference tasks.id. Unknown dates/URLs stay NULL.
type CompetitionRecord struct {
	TaskID               string     `db:"task_id"`
	OfficialURL          *string    `db:"official_url"`
	RegistrationDeadline *time.Time `db:"registration_deadline"`
	StartsAt             *time.Time `db:"starts_at"`
	Eligibility          string     `db:"eligibility"`
	TeamRules            string     `db:"team_rules"`
	VerifiedAt           *time.Time `db:"verified_at"`
}

type ProjectRecord struct {
	TaskID           string   `db:"task_id"`
	RepositoryURL    *string  `db:"repository_url"`
	License          *string  `db:"license"`
	TechStack        []string `db:"tech_stack"`
	AnalyzedRevision *string  `db:"analyzed_revision"`
}

type IssueRecord struct {
	TaskID     string    `db:"task_id"`
	Repository string    `db:"repository"` // owner/repo
	Number     int64     `db:"number"`
	URL        string    `db:"url"`
	Labels     []string  `db:"labels"`
	State      string    `db:"state"`
	Assigned   bool      `db:"assigned"`
	FetchedAt  time.Time `db:"fetched_at"`
}
