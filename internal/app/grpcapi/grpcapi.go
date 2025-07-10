package grpcapi

import (
	context "context"

	middleware "github.com/nasik90/url-shortener/internal/app/middlewares"
)

// Service - интерфейс, который описывает методы объектов с типом Service
type Service interface {
	GetShortURL(ctx context.Context, originalURL, userID string) (string, error)
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	GetShortURLs(ctx context.Context, originalURLs map[string]string, userID string) (map[string]string, error)
	GetUserURLs(ctx context.Context, userID string) (map[string]string, error)
	MarkRecordsForDeletion(ctx context.Context, shortURLs []string, userID string)
	Ping(ctx context.Context) error
	GetURLsStats(ctx context.Context) (int, int, error)
}

// ShortenerServerStruct поддерживает все необходимые методы сервера.
type ShortenerServerStruct struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	UnimplementedShortenerServer
	// service Service
	service Service
}

func NewShortenerServer(service Service) *ShortenerServerStruct {
	return &ShortenerServerStruct{service: service}
}

// GetShortURL - метод для получения короткого URL по переданному оригинальному URL.
func (s *ShortenerServerStruct) GetShortURL(ctx context.Context, req *GetShortURLRequest) (*GetShortURLResponse, error) {
	var (
		response GetShortURLResponse
		err      error
	)
	response.ShortURL, err = s.service.GetShortURL(ctx, req.OriginalURL, middleware.UserIDFromContext(ctx))
	return &response, err
}

// GetOriginalURL - метод для получения оригинального URL по переданному короткому URL.
func (s *ShortenerServerStruct) GetOriginalURL(ctx context.Context, req *GetOriginalURLRequest) (*GetOriginalURLResponse, error) {
	var (
		response GetOriginalURLResponse
		err      error
	)
	response.OriginalURL, err = s.service.GetOriginalURL(ctx, req.ShortURL)
	return &response, err
}

// GetShortURLs - принимает на вход массив структур с указанием correlation_id и оригинального URL.
// Возвращает массив струкур с указанием correlation_id и короткого URL.
func (s *ShortenerServerStruct) GetShortURLs(ctx context.Context, req *GetShortURLsRequest) (*GetShortURLsResponse, error) {
	var (
		response GetShortURLsResponse
		err      error
	)
	originalURLs := make(map[string]string)
	for _, in := range req.OriginalURLs {
		originalURLs[in.CorrelationID] = in.OriginalURL
	}
	shortURLs, err := s.service.GetShortURLs(ctx, originalURLs, middleware.UserIDFromContext(ctx))
	if err != nil {
		return &response, err
	}
	for corID, shortURL := range shortURLs {
		var shortURLWithId ShortURLWithId
		shortURLWithId.ShortURL = shortURL
		shortURLWithId.CorrelationID = corID
		response.ShortURLs = append(response.ShortURLs, &shortURLWithId)
	}
	return &response, err
}

// GetUserURLs - возвращает список URL`ов пользователя.
// Список представляет собой массив структур с указанием короткого и оригинального URL.
func (s *ShortenerServerStruct) GetUserURLs(ctx context.Context, req *GetUserURLsRequest) (*GetUserURLsResponse, error) {
	var (
		response GetUserURLsResponse
		err      error
	)
	userURLs, err := s.service.GetUserURLs(ctx, middleware.UserIDFromContext(ctx))
	if err != nil {
		return &response, err
	}
	for shortURL, originalURL := range userURLs {
		var shortOriginalURL ShortOriginalURL
		shortOriginalURL.ShortURL = shortURL
		shortOriginalURL.OriginalURL = originalURL
		response.ShortOriginalURLs = append(response.ShortOriginalURLs, &shortOriginalURL)
	}
	return &response, nil
}

// MarkRecordsForDeletion помечает на удаление переданные в массиве короткие URL
func (s *ShortenerServerStruct) MarkRecordsForDeletion(ctx context.Context, req *MarkRecordsForDeletionRequest) (*MarkRecordsForDeletionResponse, error) {
	s.service.MarkRecordsForDeletion(ctx, req.ShortURLs, middleware.UserIDFromContext(ctx))
	return nil, nil
}

// Ping - проверяет работоспособность сервера и БД.
func (s *ShortenerServerStruct) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	err := s.service.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// GetUserURLs - возвращает количество URL и пользователей.
// В настройках сервиса обязательно должен быть указан CIDR и передан в метаданных X-Real-IP IP адрес.
func (s *ShortenerServerStruct) GetURLsStats(ctx context.Context, req *GetURLsStatsRequest) (*GetURLsStatsResponse, error) {
	var (
		response GetURLsStatsResponse
		err      error
	)
	urls, users, err := s.service.GetURLsStats(ctx)
	if err != nil {
		return &response, err
	}
	response.Urls = int64(urls)
	response.Users = int64(users)
	return &response, nil
}
