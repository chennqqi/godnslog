package models

import (
	"time"

	"github.com/chennqqi/godnslog/models"
)

// Resolve represents DNS resolution configuration
// This is a wrapper around models.TblResolve for compatibility
type Resolve struct {
	ID         int64     `json:"id" xorm:"pk autoincr"`
	Host       string    `json:"host" xorm:"varchar(255) notnull index"` // host record, eg. www
	Type       string    `json:"type" xorm:"varchar(16) notnull"` // record type, eg. CNAME/A/MX/TXT/SRV/NS
	Value      string    `json:"value" xorm:"varchar(255) notnull"`
	TTL        uint32    `json:"ttl" xorm:"default 300"`
	CreatedAt  time.Time `json:"created_at" xorm:"datetime created"`
	UpdatedAt  time.Time `json:"updated_at" xorm:"datetime updated"`
}

// TableName returns the table name for Resolve model
func (Resolve) TableName() string {
	return "tbl_resolve"
}

// ToTblResolve converts Resolve to models.TblResolve
func (r *Resolve) ToTblResolve() *models.TblResolve {
	return &models.TblResolve{
		Id:    r.ID,
		Host:  r.Host,
		Type:  r.Type,
		Value: r.Value,
		Ttl:   r.TTL,
		Ctime: r.CreatedAt,
		Utime: r.UpdatedAt,
	}
}

// FromTblResolve converts models.TblResolve to Resolve
func FromTblResolve(tbl *models.TblResolve) *Resolve {
	return &Resolve{
		ID:        tbl.Id,
		Host:      tbl.Host,
		Type:      tbl.Type,
		Value:     tbl.Value,
		TTL:       tbl.Ttl,
		CreatedAt: tbl.Ctime,
		UpdatedAt: tbl.Utime,
	}
}
