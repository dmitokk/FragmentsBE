package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dmitokk/FragmentsBE/internal/application/dto"
	"github.com/dmitokk/FragmentsBE/internal/domain/entity"
	"github.com/dmitokk/FragmentsBE/internal/domain/repository"
	"github.com/dmitokk/FragmentsBE/internal/infrastructure/storage/minio"
	"github.com/google/uuid"
)

const filesBasePath = "/api/files/"

type FragmentService struct {
	fragmentRepo       repository.FragmentRepository
	userFragmentRepo   repository.UserFragmentRepository
	achievementService *AchievementService
	minioClient        *minio.Client
}

func NewFragmentService(
	fragmentRepo repository.FragmentRepository,
	userFragmentRepo repository.UserFragmentRepository,
	achievementService *AchievementService,
	minioClient *minio.Client,
) *FragmentService {
	return &FragmentService{
		fragmentRepo:       fragmentRepo,
		userFragmentRepo:   userFragmentRepo,
		achievementService: achievementService,
		minioClient:        minioClient,
	}
}

func (s *FragmentService) Create(ctx context.Context, userID uuid.UUID, req *dto.CreateFragmentRequest, photos []io.Reader, photoSizes []int64, sound io.Reader, soundSize int64) (*dto.FragmentResponse, error) {
	fragmentID := uuid.New()

	var photoURLs []string
	for i, photo := range photos {
		filename := fmt.Sprintf("photo_%d.jpg", i)
		objectName, err := s.minioClient.UploadFile(ctx, fragmentID.String(), "photos", filename, photo, photoSizes[i])
		if err != nil {
			return nil, fmt.Errorf("failed to upload photo: %w", err)
		}
		photoURLs = append(photoURLs, objectName)
	}

	var soundURL string
	if sound != nil {
		filename := "sound.mp3"
		objectName, err := s.minioClient.UploadFile(ctx, fragmentID.String(), "sounds", filename, sound, soundSize)
		if err != nil {
			return nil, fmt.Errorf("failed to upload sound: %w", err)
		}
		soundURL = objectName
	}

	fragment := &entity.Fragment{
		ID:        fragmentID,
		UserID:    userID,
		Text:      req.Text,
		Geomark:   &entity.Geomark{Lat: *req.Lat, Lng: *req.Lng},
		SoundURL:  soundURL,
		PhotoURLs: photoURLs,
		ExpiresAt: time.Now().Add(time.Duration(req.GetLifetimeHours()) * time.Hour),
	}

	err := s.fragmentRepo.Create(ctx, fragment)
	if err != nil {
		return nil, fmt.Errorf("failed to create fragment: %w", err)
	}

	return s.toResponse(fragment), nil
}

func (s *FragmentService) GetByID(ctx context.Context, id uuid.UUID) (*dto.FragmentResponse, error) {
	fragment, err := s.fragmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(fragment), nil
}

func (s *FragmentService) List(ctx context.Context, userID uuid.UUID, lat, lng, radius float64) ([]*dto.FragmentResponse, error) {
	fragments, err := s.fragmentRepo.List(ctx, userID, lat, lng, radius)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.FragmentResponse, len(fragments))
	for i, fragment := range fragments {
		responses[i] = s.toResponse(fragment)
	}

	return responses, nil
}

func (s *FragmentService) MarkFound(ctx context.Context, userID uuid.UUID, fragmentID uuid.UUID) error {
	exists, err := s.userFragmentRepo.Exists(ctx, userID, fragmentID)
	if err != nil {
		return fmt.Errorf("failed to check existing fragment: %w", err)
	}
	if exists {
		return nil
	}

	err = s.userFragmentRepo.Create(ctx, userID, fragmentID)
	if err != nil {
		return fmt.Errorf("failed to mark fragment as found: %w", err)
	}

	if s.achievementService != nil {
		if err := s.achievementService.CheckAndUnlock(ctx, userID, fragmentID); err != nil {
			return fmt.Errorf("failed to check achievements: %w", err)
		}
	}

	return nil
}

func (s *FragmentService) GetFound(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	userFragments, err := s.userFragmentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, len(userFragments))
	for i, uf := range userFragments {
		ids[i] = uf.FragmentID
	}

	return ids, nil
}

func (s *FragmentService) toResponse(fragment *entity.Fragment) *dto.FragmentResponse {
	response := &dto.FragmentResponse{
		ID:        fragment.ID,
		UserID:    fragment.UserID,
		Text:      fragment.Text,
		SoundURL:  resolveFileURL(fragment.SoundURL),
		PhotoURLs: resolveFileURLs(fragment.PhotoURLs),
		ExpiresAt: fragment.ExpiresAt.Format(time.RFC3339),
		CreatedAt: fragment.CreatedAt.Format(time.RFC3339),
		UpdatedAt: fragment.UpdatedAt.Format(time.RFC3339),
	}

	if fragment.Geomark != nil {
		response.Lat = fragment.Geomark.Lat
		response.Lng = fragment.Geomark.Lng
	}

	return response
}

func resolveFileURL(url string) string {
	if url == "" || strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return filesBasePath + url
}

func resolveFileURLs(urls []string) []string {
	if urls == nil {
		return nil
	}
	resolved := make([]string, len(urls))
	for i, u := range urls {
		resolved[i] = resolveFileURL(u)
	}
	return resolved
}