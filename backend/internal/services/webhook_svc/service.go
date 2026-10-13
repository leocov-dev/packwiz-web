package webhook_svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/riverqueue/river"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"packwiz-web/internal/config"
	"packwiz-web/internal/jobs"
	"packwiz-web/internal/log"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
	"packwiz-web/internal/utils"
)

// Delay is how far in the future a webhook call is scheduled, so a burst of
// edits collapses into one call (the job is unique per pack).
const Delay = 15 * time.Second

// ScanRequester is the pack-change hook of packwiz_svc and import_svc, which
// each declare the same single-method interface.
type ScanRequester interface {
	RequestScan(ctx context.Context, packID uint) error
}

// Service stores the webhook of a pack and calls it when the pack changes.
type Service struct {
	db *gorm.DB
	// riverClient enqueues calls. It is nil for job resolver instances, which
	// enqueue through the client River puts in a worker context.
	riverClient   *river.Client[*sql.Tx]
	sessionSecret []byte
	http          *http.Client
	now           func() time.Time
}

// NewService builds the service. riverClient may be nil (job resolver
// instances). sessionSecret keys the encryption of the stored headers.
func NewService(db *gorm.DB, riverClient *river.Client[*sql.Tx], sessionSecret []byte) *Service {
	return &Service{db: db, riverClient: riverClient, sessionSecret: sessionSecret, http: newHTTPClient(), now: time.Now}
}

func (s *Service) client(ctx context.Context) (*river.Client[*sql.Tx], error) {
	if s.riverClient != nil {
		return s.riverClient, nil
	}
	if c, err := river.ClientFromContextSafely[*sql.Tx](ctx); err == nil {
		return c, nil
	}
	return nil, jobs.ErrClientUnavailable
}

// RequestPackChanged schedules the webhook call of a pack, Delay from now,
// when the pack has an enabled webhook. Repeated calls while one is pending
// collapse into it. Call it after the change has committed.
func (s *Service) RequestPackChanged(ctx context.Context, packID uint) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&tables.PackWebhook{}).
		Where("pack_id = ? AND enabled AND url <> ''", packID).Count(&count).Error; err != nil {
		return fmt.Errorf("look up webhook of pack %d: %w", packID, err)
	}
	if count == 0 {
		return nil
	}
	c, err := s.client(ctx)
	if err != nil {
		return err
	}
	_, err = c.Insert(ctx, jobs.PackWebhookArgs{PackID: packID}, &river.InsertOpts{ScheduledAt: s.now().Add(Delay)})
	if err != nil {
		return fmt.Errorf("enqueue webhook of pack %d: %w", packID, err)
	}
	return nil
}

// chain forwards scan requests to next and also requests the webhook call.
type chain struct {
	next ScanRequester
	svc  *Service
}

func (c chain) RequestScan(ctx context.Context, packID uint) error {
	var scanErr error
	if c.next != nil {
		scanErr = c.next.RequestScan(ctx, packID)
	}
	// a failing scan request must not suppress the webhook, and vice versa
	return errors.Join(scanErr, c.svc.RequestPackChanged(ctx, packID))
}

// Chain returns a ScanRequester that forwards to next (may be nil) and then
// requests the webhook call. The packwiz and import services request a scan
// after every committed pack change, so this is the one hook for all of them.
func (s *Service) Chain(next ScanRequester) ScanRequester {
	return chain{next: next, svc: s}
}

// -----------------------------------------------------------------------------

func (s *Service) usingDefaultSecret() bool {
	return string(s.sessionSecret) == config.DefaultSessionSecret
}

func (s *Service) load(ctx context.Context, packID uint) (*tables.PackWebhook, error) {
	var hook tables.PackWebhook
	err := s.db.WithContext(ctx).First(&hook, "pack_id = ?", packID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load webhook of pack %d: %w", packID, err)
	}
	return &hook, nil
}

// decryptHeaders returns the stored headers. ok is false when they cannot be
// decrypted or parsed.
func (s *Service) decryptHeaders(hook *tables.PackWebhook) (headers []header, ok bool) {
	if hook == nil || hook.HeadersEnc == "" {
		return nil, true
	}
	plain, err := utils.SecretboxDecrypt(s.sessionSecret, hook.HeadersEnc)
	if err != nil {
		return nil, false
	}
	if err := json.Unmarshal([]byte(plain), &headers); err != nil {
		return nil, false
	}
	return headers, true
}

func (s *Service) toResponse(hook *tables.PackWebhook) dto.PackWebhookResponse {
	if hook == nil {
		return dto.PackWebhookResponse{Method: http.MethodPost, Headers: []dto.WebhookHeaderResponse{}}
	}
	headers, ok := s.decryptHeaders(hook)
	names := make([]dto.WebhookHeaderResponse, 0, len(headers))
	for _, h := range headers {
		names = append(names, dto.WebhookHeaderResponse{Name: h.Name})
	}
	return dto.PackWebhookResponse{
		Enabled:        hook.Enabled,
		URL:            hook.URL,
		Method:         hook.Method,
		Headers:        names,
		Broken:         !ok,
		LastStatusCode: hook.LastStatusCode,
		LastError:      hook.LastError,
		LastAttemptAt:  hook.LastAttemptAt,
		LastSuccessAt:  hook.LastSuccessAt,
	}
}

// Get returns the webhook of a pack; a pack without one yields a disabled,
// empty webhook.
func (s *Service) Get(ctx context.Context, packID uint) (dto.PackWebhookResponse, response.ServerError) {
	hook, err := s.load(ctx, packID)
	if err != nil {
		return dto.PackWebhookResponse{}, response.Wrap(err)
	}
	return s.toResponse(hook), nil
}

// mergeHeaders builds the stored header set from the form: a blank value keeps
// the existing value of the header with that name.
func mergeHeaders(existing []header, form []dto.WebhookHeaderForm) ([]header, error) {
	old := make(map[string]string, len(existing))
	for _, h := range existing {
		old[strings.ToLower(h.Name)] = h.Value
	}
	out := make([]header, 0, len(form))
	for _, f := range form {
		value := f.Value
		if value == "" {
			kept, ok := old[strings.ToLower(f.Name)]
			if !ok {
				return nil, fmt.Errorf("header %q needs a value", f.Name)
			}
			value = kept
		}
		out = append(out, header{Name: f.Name, Value: value})
	}
	return out, nil
}

// Save replaces the webhook of a pack.
func (s *Service) Save(ctx context.Context, packID uint, form dto.PackWebhookForm) (dto.PackWebhookResponse, response.ServerError) {
	if len(form.Headers) > 0 && s.usingDefaultSecret() {
		return dto.PackWebhookResponse{}, response.New(
			http.StatusBadRequest,
			"webhook headers cannot be saved while the default session secret is in use; set PWW_SESSION_SECRET to a private random value and restart",
		)
	}

	hook, err := s.load(ctx, packID)
	if err != nil {
		return dto.PackWebhookResponse{}, response.Wrap(err)
	}
	// headers that cannot be decrypted are treated as absent, so the user can
	// re-enter them
	existing, _ := s.decryptHeaders(hook)
	merged, err := mergeHeaders(existing, form.Headers)
	if err != nil {
		return dto.PackWebhookResponse{}, response.New(http.StatusBadRequest, err.Error())
	}

	enc := ""
	if len(merged) > 0 {
		raw, err := json.Marshal(merged)
		if err != nil {
			return dto.PackWebhookResponse{}, response.Wrap(err)
		}
		enc, err = utils.SecretboxEncrypt(s.sessionSecret, string(raw))
		if err != nil {
			return dto.PackWebhookResponse{}, response.New(http.StatusInternalServerError, "failed to encrypt webhook headers")
		}
	}

	row := tables.PackWebhook{
		PackID:     packID,
		Enabled:    form.Enabled,
		URL:        form.URL,
		Method:     form.Method,
		HeadersEnc: enc,
	}
	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pack_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"enabled", "url", "method", "headers_enc", "updated_at"}),
	}).Create(&row).Error; err != nil {
		return dto.PackWebhookResponse{}, response.New(http.StatusInternalServerError, "failed to save webhook")
	}

	saved, err := s.load(ctx, packID)
	if err != nil {
		return dto.PackWebhookResponse{}, response.Wrap(err)
	}
	return s.toResponse(saved), nil
}

// -----------------------------------------------------------------------------

// deliver calls the webhook of a pack with the event and records the outcome.
func (s *Service) deliver(ctx context.Context, hook *tables.PackWebhook, event string) (outcome, error) {
	var pack tables.Pack
	if err := s.db.WithContext(ctx).Select("id", "slug", "name", "head_snapshot_id").
		First(&pack, "id = ?", hook.PackID).Error; err != nil {
		return outcome{}, fmt.Errorf("load pack %d: %w", hook.PackID, err)
	}
	headers, ok := s.decryptHeaders(hook)
	if !ok {
		out := outcome{Err: "stored headers cannot be decrypted; enter them again"}
		s.record(ctx, hook.PackID, out)
		return out, nil
	}
	body := payload{
		Event:          event,
		PackID:         pack.ID,
		Slug:           pack.Slug,
		Name:           pack.Name,
		HeadSnapshotID: pack.HeadSnapshotID,
		At:             s.now().UTC().Format(time.RFC3339),
	}
	out := send(ctx, s.http, hook.Method, hook.URL, headers, body)
	s.record(ctx, hook.PackID, out)
	return out, nil
}

// record stores the outcome of an attempt. Failing to do so is only logged.
func (s *Service) record(ctx context.Context, packID uint, out outcome) {
	now := s.now()
	updates := map[string]any{
		"last_status_code": out.StatusCode,
		"last_error":       out.Err,
		"last_attempt_at":  now,
	}
	if out.ok() {
		updates["last_success_at"] = now
	}
	// context.WithoutCancel: the status should be stored even when the call
	// was cut short by the job timeout
	if err := s.db.WithContext(context.WithoutCancel(ctx)).Model(&tables.PackWebhook{}).
		Where("pack_id = ?", packID).Updates(updates).Error; err != nil {
		log.Error(fmt.Errorf("record webhook outcome of pack %d: %w", packID, err))
	}
}

// RunWebhook is the job body: it calls the webhook if it is still enabled. An
// error is returned only for failures worth retrying.
func (s *Service) RunWebhook(ctx context.Context, args jobs.PackWebhookArgs) error {
	hook, err := s.load(ctx, args.PackID)
	if err != nil {
		return err
	}
	if hook == nil || !hook.Enabled || hook.URL == "" {
		return nil
	}
	out, err := s.deliver(ctx, hook, "pack.changed")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil // pack archived or deleted since
	}
	if err != nil {
		return err
	}
	if !out.ok() && out.Retry {
		return errors.New(out.Err)
	}
	return nil
}

// Test calls the saved webhook right away, enabled or not, and records the
// outcome.
func (s *Service) Test(ctx context.Context, packID uint) (dto.PackWebhookTestResponse, response.ServerError) {
	hook, err := s.load(ctx, packID)
	if err != nil {
		return dto.PackWebhookTestResponse{}, response.Wrap(err)
	}
	if hook == nil || hook.URL == "" {
		return dto.PackWebhookTestResponse{}, response.New(http.StatusBadRequest, "save a webhook URL first")
	}
	out, err := s.deliver(ctx, hook, "pack.test")
	if err != nil {
		return dto.PackWebhookTestResponse{}, response.Wrap(err)
	}
	return dto.PackWebhookTestResponse{OK: out.ok(), StatusCode: out.StatusCode, Error: out.Err}, nil
}
