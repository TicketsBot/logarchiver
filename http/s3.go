package http

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v6"
)

var ErrTicketNotFound = errors.New("object not found")

// data, is_premium, error
func (s *Server) GetTicket(bucketName string, guildId uint64, ticketId int) ([]byte, bool, error) {
	// try reading with free name
	reader, err := s.client.GetObject(bucketName, fmt.Sprintf("%d/free-%d", guildId, ticketId), minio.GetObjectOptions{})
	if err != nil {
		return nil, false, err
	}

	// if we found the free object, we can return it
	if reader != nil {
		defer reader.Close()

		var buff bytes.Buffer
		_, err = buff.ReadFrom(reader)
		if err != nil {
			if err.Error() != "The specified key does not exist." {
				return nil, false, err
			}
		} else {
			return buff.Bytes(), false, nil
		}
	}

	// else, we should check the premium object
	reader, err = s.client.GetObject(bucketName, fmt.Sprintf("%d/%d", guildId, ticketId), minio.GetObjectOptions{})
	if err != nil {
		return nil, false, err
	}

	if reader != nil {
		defer reader.Close()

		var buff bytes.Buffer
		_, err = buff.ReadFrom(reader)
		if err != nil {
			if err.Error() != "The specified key does not exist." {
				return nil, false, err
			}
		} else {
			return buff.Bytes(), true, nil
		}
	}

	return nil, false, ErrTicketNotFound
}

func (s *Server) UploadTicket(bucket string, isPremium bool, guildId uint64, ticketId int, data []byte) error {
	var freePrefix string
	if !isPremium {
		freePrefix = "free-"
	}

	name := fmt.Sprintf("%d/%s%d", guildId, freePrefix, ticketId)

	_, err := s.client.PutObject(bucket, name, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType:     "application/octet-stream",
		ContentEncoding: "zstd",
	})

	return err
}
