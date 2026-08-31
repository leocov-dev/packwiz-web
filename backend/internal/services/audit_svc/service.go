package audit_svc

import (
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
	"time"
)

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) ListAudits(request dto.ListAuditsQuery) ([]tables.Audit, int64, response.ServerError) {
	var audits []tables.Audit
	var total int64

	offset := (request.Page - 1) * request.PageSize

	if err := s.db.Transaction(func(tx *gorm.DB) error {

		query := tx.Model(&tables.Audit{})

		if request.Action != "" {
			query.Where("action ILIKE ?", "%"+request.Action+"%")
		}

		if request.UserId > 0 {
			query.Where("user_id = ?", request.UserId)
		}

		if request.StartDate != "" {
			if start, err := time.Parse("2006-01-02", request.StartDate); err == nil {
				query.Where("created_at >= ?", start)
			}
		}

		if request.EndDate != "" {
			if end, err := time.Parse("2006-01-02", request.EndDate); err == nil {
				query.Where("created_at < ?", end.AddDate(0, 0, 1))
			}
		}

		if err := query.Count(&total).Error; err != nil {
			return err
		}

		if err := query.Order("created_at DESC").Offset(offset).Limit(request.PageSize).Scan(&audits).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, 0, response.New(http.StatusInternalServerError, "failed to list audits")
	}

	if audits == nil {
		audits = []tables.Audit{}
	}

	return audits, total, nil
}
