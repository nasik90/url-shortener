package service

import (
	"context"
	"time"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/logger"
	"go.uber.org/zap"
)

// HandleRecords помечает на удаление короткие урлы, которые находятся в канале recordsForDel.
func (s *Service) HandleRecords() {
	// будем сохранять сообщения, накопленные за последние 5 секунд
	ticker := time.NewTicker(5 * time.Second)

	var records []settings.Record
	for {
		select {
		case record := <-s.recordsForDel:
			// добавим сообщение в слайс для последующего сохранения
			records = append(records, record)
		case <-ticker.C:
			// подождём, пока придёт хотя бы одно сообщение
			if len(records) == 0 {
				continue
			}
			// сохраним все пришедшие сообщения одновременно
			err := s.repo.MarkRecordsForDeletion(context.TODO(), records...)
			if err != nil {
				logger.Log.Info("cannot mark records for deletion", zap.Error(err))
				// не будем стирать сообщения, попробуем отправить их чуть позже
				continue
			}
			// сотрём успешно отосланные сообщения
			records = nil
		}
	}
}
