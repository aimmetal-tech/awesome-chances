package model

import "time"

// MigrationRecord 仅供版本迁移使用；AppliedAt 由 PostgreSQL 默认值生成，只读。
type MigrationRecord struct {
	Version   string    `gorm:"column:version;primaryKey"`
	Checksum  string    `gorm:"column:checksum"`
	AppliedAt time.Time `gorm:"column:applied_at;->"`
}

func (MigrationRecord) TableName() string { return "schema_migrations" }
