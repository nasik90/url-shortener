package storage

import (
	"bufio"
	"encoding/json"
	"os"
)

// Event - структура для хранения данных в json в файле.
type Event struct {
	UUID         string `json:"uuid"`
	ShortURL     string `json:"short_url"`
	OriginalURL  string `json:"original_url"`
	UserID       string `json:"user_id"`
	MarkedForDel bool   `json:"del"`
}

// Producer - структура для хранения данных о писателе в файл.
type Producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}

// NewProducer создает экземпляр типа Producer.
func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file: file,
		// создаём новый Writer
		writer: bufio.NewWriter(file),
	}, nil
}

// WriteEvent пишет в файл событие типа Event.
func (p *Producer) WriteEvent(event *Event) error {
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return p.writer.Flush()
}

// Close закрывает писателя в файл.
func (p *Producer) Close() error {
	// закрываем файл
	if err := p.writer.Flush(); err != nil {
		return err
	}
	return p.file.Close()
}

// Consumer - структура, хранящая данные о читателя из файла.
type Consumer struct {
	file *os.File
	// добавляем reader в Consumer
	reader *bufio.Reader
}

// NewConsumer создает экземпляр типа Consumer.
func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file: file,
		// создаём новый Reader
		reader: bufio.NewReader(file),
	}, nil
}

// ReadEvent читает из файла данные и возвращает струтуру типа Event.
func (c *Consumer) ReadEvent() (*Event, error) {
	// читаем данные до символа переноса строки
	data, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	// преобразуем данные из JSON-представления в структуру
	event := Event{}
	err = json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

// Close закрывает читателя фай
func (c *Consumer) Close() error {
	// закрываем файл
	return c.file.Close()
}
