package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/dmitokk/FragmentsBE/internal/application/dto"
	"github.com/dmitokk/FragmentsBE/internal/domain/entity"
	"github.com/dmitokk/FragmentsBE/internal/domain/repository"
	"github.com/dmitokk/FragmentsBE/internal/infrastructure/storage/minio"
	"github.com/google/uuid"
)

type FragmentService struct {
	fragmentRepo repository.FragmentRepository
	minioClient  *minio.Client
}

func NewFragmentService(fragmentRepo repository.FragmentRepository, minioClient *minio.Client) *FragmentService {
	return &FragmentService{
		fragmentRepo: fragmentRepo,
		minioClient:  minioClient,
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

func (s *FragmentService) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateFragmentRequest) (*dto.FragmentResponse, error) {
	fragment, err := s.fragmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Text != "" {
		fragment.Text = req.Text
	}

	if req.Lat != 0 || req.Lng != 0 {
		fragment.Geomark = &entity.Geomark{Lat: req.Lat, Lng: req.Lng}
	}

	err = s.fragmentRepo.Update(ctx, fragment)
	if err != nil {
		return nil, err
	}

	return s.toResponse(fragment), nil
}

func (s *FragmentService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.minioClient.DeleteFiles(ctx, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete files: %w", err)
	}

	return s.fragmentRepo.Delete(ctx, id)
}

func (s *FragmentService) toResponse(fragment *entity.Fragment) *dto.FragmentResponse {
	response := &dto.FragmentResponse{
		ID:        fragment.ID,
		UserID:    fragment.UserID,
		Text:      fragment.Text,
		SoundURL:  fragment.SoundURL,
		PhotoURLs: fragment.PhotoURLs,
		CreatedAt: fragment.CreatedAt.Format(time.RFC3339),
		UpdatedAt: fragment.UpdatedAt.Format(time.RFC3339),
	}

	if fragment.Geomark != nil {
		response.Lat = fragment.Geomark.Lat
		response.Lng = fragment.Geomark.Lng
	}

	return response
}