package http

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v6"
)

var ErrTicketNotFound = errors.New("object not found")

// data, error
func (s *Server) GetTicket(bucketName string, guildId uint64, ticketId int) ([]byte, error) {
	// try reading with free name
	reader, err := s.client.GetObject(bucketName, fmt.Sprintf("%d/free-%d", guildId, ticketId), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	// if we found the free object, we can return it
	if reader != nil {
		defer reader.Close()

		var buff bytes.Buffer
		_, err = buff.ReadFrom(reader)
		if err != nil {
			if err.Error() != "The specified key does not exist." {
				return nil, err
			}
		} else {
			return buff.Bytes(), nil
		}
	}

	// else, we should check the premium object
	reader, err = s.client.GetObject(bucketName, fmt.Sprintf("%d/%d", guildId, ticketId), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	if reader != nil {
		defer reader.Close()

		var buff bytes.Buffer
		_, err = buff.ReadFrom(reader)
		if err != nil {
			if err.Error() != "The specified key does not exist." {
				return nil, err
			}
		} else {
			return buff.Bytes(), nil
		}
	}

	return nil, ErrTicketNotFound
}

func (s *Server) UploadTicket(bucket string, guildId uint64, ticketId int, data []byte) error {
	name := fmt.Sprintf("%d/%d", guildId, ticketId)

	_, err := s.client.PutObject(bucket, name, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType:     "application/octet-stream",
		ContentEncoding: "zstd",
	})

	return err
}
