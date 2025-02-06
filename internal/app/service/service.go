package service

import (
	"crypto/rand"
	"math/big"
	"sync"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

type repositories interface {
	SaveShortURL(shortURL, originalURL string) error
	GetOriginalURL(shortURL string) (string, error)
	ShortURLUnique(shortURL string) bool
}

// var (
// 	URLWriterToFile *storage.Producer
// 	CurrentUUID     int
// )

func GetShortURL(repository repositories, mutex *sync.Mutex, originalURL, host string) (string, error) {
	mutex.Lock()
	shortURL, err := shortURLWithRetrying(repository)
	if err != nil {
		return "", err
	}
	repository.SaveShortURL(shortURL, originalURL)
	// saveShortURLToFile(shortURL, originalURL)
	mutex.Unlock()

	shortURLWithHost := shortURLWithHost(host, shortURL)
	return shortURLWithHost, nil
}

// func saveShortURLToFile(shortURL string, originalURL string) {
// 	var event storage.Event
// 	CurrentUUID++
// 	event.UUID = strconv.Itoa(CurrentUUID)
// 	event.ShortURL = shortURL
// 	event.OriginalURL = originalURL
// 	URLWriterToFile.WriteEvent(&event)
// }

func GetOriginalURL(repository repositories, shortURL string) (string, error) {

	originalURL, err := repository.GetOriginalURL(shortURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil

}

func randomString(charCount int) (res string, err error) {
	template := []rune(settings.TemplateForRand)
	templateLen := len(template)
	resChar := make([]rune, charCount)
	for i := 0; i < charCount; i++ {

		r, err := rand.Int(rand.Reader, big.NewInt(int64(templateLen)))
		if err != nil {
			return res, err
		}
		resChar[i] = template[int(r.Int64())]
	}

	return string(resChar), nil
}

func shortURLWithHost(host, randomString string) string {
	// return "http://" + host + "/" + randomString
	return host + "/" + randomString
}

func shortURLWithRetrying(repository repositories) (string, error) {
	shortURL := ""
	shortURLUnique := false
	for !shortURLUnique {
		randomString, err := randomString(settings.ShortURLlen)
		if err != nil {
			return "", err
		}
		//shortURL = buildShortURL(host, randomString)
		shortURL = randomString
		shortURLUnique = repository.ShortURLUnique(shortURL)
	}
	return shortURL, nil
}

// func RestoreData(repository repositories, filePath string) error {
// 	consumer, err := storage.NewConsumer(filePath)
// 	if err != nil {
// 		return err
// 	}
// 	for {
// 		event, err := consumer.ReadEvent()
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			return err
// 		}
// 		err = repository.SaveShortURL(event.ShortURL, event.OriginalURL)
// 		if err != nil {
// 			return err
// 		}
// 		CurrentUUID, err = strconv.Atoi(event.UUID)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	consumer.Close()
// 	return nil
// }

// func NewProducer(FilePath string) (*storage.Producer, error) {
// 	return storage.NewProducer(FilePath)
// }

// func DestroyProducer(producer *storage.Producer) error {
// 	return producer.Close()
// }
