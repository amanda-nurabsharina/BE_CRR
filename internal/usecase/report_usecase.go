package usecase

import (
	"context"
	"fmt"
	"time"

	"be_crr/internal/domain"
	"be_crr/internal/domain/entity"
	"be_crr/internal/domain/repository"
	"be_crr/internal/infrastructure/hasher"
	"be_crr/internal/infrastructure/storage"

	"github.com/google/uuid"
)

type ReportUsecase struct {
	reportRepo  repository.ReportRepository
	sessionRepo repository.SessionRepository
	audioRepo   repository.AudioRepository
	storage     storage.Storage
	auditRepo   repository.AuditRepository
}

func NewReportUsecase(
	reportRepo repository.ReportRepository,
	sessionRepo repository.SessionRepository,
	audioRepo repository.AudioRepository,
	storage storage.Storage,
	auditRepo repository.AuditRepository,
) *ReportUsecase {
	return &ReportUsecase{
		reportRepo:  reportRepo,
		sessionRepo: sessionRepo,
		audioRepo:   audioRepo,
		storage:     storage,
		auditRepo:   auditRepo,
	}
}

func (u *ReportUsecase) CreateReport(ctx context.Context, tenantID, sessionID uuid.UUID, title string) (*entity.CourtReport, error) {
	sess, err := u.sessionRepo.GetByID(ctx, tenantID, sessionID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	report := &entity.CourtReport{
		TenantID:           tenantID,
		CourtSessionID:     sessionID,
		ReportTitle:        title,
		DocumentNumber:     fmt.Sprintf("BAS/%s/%d", sess.SessionDate.Format("20060102"), time.Now().Unix()%1000),
		CaseSummary:        "Berita acara pencatatan jalannya persidangan dan keterangan para pihak serta saksi di ruang sidang.",
		KeyTestimoniesJSON: "[]",
		Status:             entity.ReportStatusDraft,
		Version:            1,
	}
	if report.ReportTitle == "" {
		report.ReportTitle = "Berita Acara Jalannya Persidangan"
	}

	if err := u.reportRepo.Create(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

func (u *ReportUsecase) GetReport(ctx context.Context, tenantID, reportID uuid.UUID) (*entity.CourtReport, error) {
	r, err := u.reportRepo.GetByID(ctx, tenantID, reportID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	return r, nil
}

func (u *ReportUsecase) UpdateReport(ctx context.Context, tenantID uuid.UUID, report *entity.CourtReport) error {
	report.TenantID = tenantID
	report.UpdatedAt = time.Now()
	report.Version++
	return u.reportRepo.Update(ctx, report)
}

type SignReportDTO struct {
	TenantID           uuid.UUID
	ReportID           uuid.UUID
	SignerName         string
	SignerIdentifier   string // NIA / NIK
	SignerRole         string // Advokat / Notulis
	SignatureImageData string // Data URL / Base64 canvas image
	SignerIPAddress    string
}

func (u *ReportUsecase) SignReport(ctx context.Context, dto SignReportDTO) (*entity.ReportSignature, error) {
	rep, err := u.reportRepo.GetByID(ctx, dto.TenantID, dto.ReportID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	// Retrieve session and associated audio evidence hash
	var sourceAudioHash string = "0000000000000000000000000000000000000000000000000000000000000000"
	sess, err := u.sessionRepo.GetByID(ctx, dto.TenantID, rep.CourtSessionID)
	if err == nil && len(sess.AudioEvidences) > 0 {
		sourceAudioHash = sess.AudioEvidences[0].Sha256Hash
	}

	// Compute document content hash
	docContent := fmt.Sprintf("%s|%s|%s|%s|%s", rep.DocumentNumber, rep.ReportTitle, rep.CaseSummary, rep.KeyTestimoniesJSON, rep.ObjectionsNotes)
	docHash := hasher.ComputeSHA256([]byte(docContent))

	token := fmt.Sprintf("SIG-%s-%s", time.Now().Format("20060102"), uuid.New().String()[:8])

	signature := &entity.ReportSignature{
		CourtReportID:      rep.ID,
		VerificationToken:  token,
		SignerName:         dto.SignerName,
		SignerIdentifier:   dto.SignerIdentifier,
		SignerRole:         dto.SignerRole,
		SignatureImagePath: dto.SignatureImageData,
		SourceAudioSha256:  sourceAudioHash,
		DocumentSha256:     docHash,
		SignerIPAddress:    dto.SignerIPAddress,
		SignedAt:           time.Now(),
	}

	if err := u.reportRepo.CreateSignature(ctx, signature); err != nil {
		return nil, err
	}

	rep.Status = entity.ReportStatusSigned
	rep.PdfSha256Hash = docHash
	rep.UpdatedAt = time.Now()
	_ = u.reportRepo.Update(ctx, rep)

	return signature, nil
}

type PublicVerificationResultDTO struct {
	IsValid           bool      `json:"is_valid"`
	VerificationToken string    `json:"verification_token"`
	SignerName        string    `json:"signer_name"`
	SignerRole        string    `json:"signer_role"`
	DocumentSha256    string    `json:"document_sha256"`
	SourceAudioSha256 string    `json:"source_audio_sha256"`
	SignedAt          time.Time `json:"signed_at"`
	LegalNotice       string    `json:"legal_notice"`
}

func (u *ReportUsecase) PublicVerify(ctx context.Context, token string) (*PublicVerificationResultDTO, error) {
	sig, err := u.reportRepo.GetSignatureByToken(ctx, token)
	if err != nil {
		return &PublicVerificationResultDTO{
			IsValid:     false,
			LegalNotice: "Dokumen tanda tangan tidak ditemukan atau tidak valid.",
		}, nil
	}

	return &PublicVerificationResultDTO{
		IsValid:           true,
		VerificationToken: sig.VerificationToken,
		SignerName:        sig.SignerName,
		SignerRole:        sig.SignerRole,
		DocumentSha256:    sig.DocumentSha256,
		SourceAudioSha256: sig.SourceAudioSha256,
		SignedAt:          sig.SignedAt,
		LegalNotice:       "Tanda Tangan Elektronik Tidak Tersertifikasi memenuhi Pasal 11 UU ITE & PP No. 71/2019.",
	}, nil
}

