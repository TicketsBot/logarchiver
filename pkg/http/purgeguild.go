package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/TicketsBot/logarchiver/internal"
	"github.com/TicketsBot/logarchiver/pkg/repository"
	"github.com/TicketsBot/logarchiver/pkg/repository/model"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

func (s *Server) purgeGuildHandler(ctx *gin.Context) {
	guildId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{
			"success": false,
			"message": "missing guild ID",
		})
		return
	}

	defaultClient, err := s.s3Clients.Get(s.Config.DefaultBucketId)
	if err != nil {
		s.Logger.Error("Failed to get default S3 client", zap.Error(err), zap.String("bucket_id", s.Config.DefaultBucketId.String()))

		ctx.JSON(500, gin.H{
			"success": false,
			"message": "Failed to get default S3 client",
		})
		return
	}

	if err := s.RemoveQueue.StartOperation(guildId); err != nil {
		if errors.Is(err, internal.ErrOperationInProgress) {
			ctx.JSON(400, gin.H{
				"success": false,
				"message": "operation already in progress",
			})
		} else {
			ctx.JSON(500, gin.H{
				"success": false,
				"message": err.Error(),
			})
		}

		return
	}

	type record struct {
		model.Object
		Key string
	}

	removeCh := make(chan record)
	go func() {
		var errCount int
		for r := range removeCh {
			client, err := s.s3Clients.Get(r.BucketId)
			if err != nil {
				s.Logger.Error("Failed to get S3 client", zap.Error(err), zap.String("bucket_id", r.BucketId.String()))
				continue
			}

			if err := client.Minio().RemoveObject(context.Background(), client.BucketName(), r.Key, minio.RemoveObjectOptions{}); err != nil {
				s.RemoveQueue.AddError(guildId, r.Key, err)

				s.Logger.Error(
					"Failed to remove object",
					zap.Error(err),
					zap.String("object", r.Key),
					zap.Uint64("guild", guildId),
				)

				errCount++
			}
		}

		if errCount > 0 {
			s.RemoveQueue.SetStatus(guildId, internal.StatusFailed)
			s.Logger.Warn(
				"Remove operation completed with error(s)",
				zap.Int("error_count", errCount),
				zap.Uint64("guild", guildId),
			)
		} else {
			s.RemoveQueue.SetStatus(guildId, internal.StatusComplete)
			s.Logger.Info("Remove operation completed successfully", zap.Uint64("guild", guildId))
		}
	}()

	// For the default bucket, we'll have to list all objects. For new buckets, we can fetch a list of objects from
	// the database.

	latch := sync.WaitGroup{}
	latch.Add(2)

	// Fetch from the default bucket
	go func() {
		objCh := defaultClient.Minio().ListObjects(context.Background(), defaultClient.BucketName(), minio.ListObjectsOptions{
			Prefix:    fmt.Sprintf("%d/", guildId),
			Recursive: true,
		})

		for obj := range objCh {
			s.Logger.Info(
				"Found object to remove",
				zap.String("object", obj.Key),
				zap.Uint64("guild", guildId),
			)

			// Parse ticket ID, in form guildId/ticketId or guildId/free-ticketId
			cut := obj.Key[len(fmt.Sprintf("%d/", guildId)):]
			if strings.HasPrefix(cut, "free-") {
				cut = cut[len("free-"):]
			}

			ticketId, err := strconv.Atoi(cut)
			if err != nil {
				s.Logger.Error("Failed to parse ticket ID", zap.Error(err), zap.String("object", obj.Key), zap.Uint64("guild", guildId))
				s.RemoveQueue.AddError(guildId, obj.Key, err)
				continue
			}

			s.RemoveQueue.AddRemovedObject(guildId, obj.Key)
			removeCh <- record{
				Object: model.Object{
					GuildId:  guildId,
					TicketId: ticketId,
					BucketId: s.Config.DefaultBucketId,
				},
				Key: obj.Key,
			}
		}

		close(removeCh)
	}()

	// Fetch from the database
	go func() {
		var objects []model.Object
		if err := s.store.Tx(context.Background(), func(r repository.Repositories) (err error) {
			objects, err = r.Objects().ListByGuild(context.Background(), guildId)
			return
		}); err != nil {
			s.Logger.Error("Failed to fetch objects from database", zap.Error(err), zap.Uint64("guild", guildId))
			return
		}

		for _, obj := range objects {
			s.Logger.Debug("Found object to remove", zap.String("bucket", obj.BucketId.String()), zap.Uint64("guild", guildId), zap.Int("ticket_id", obj.TicketId))
			s.RemoveQueue.AddRemovedObject(guildId, obj.S3Key())
			removeCh <- record{
				Object: obj,
				Key:    obj.S3Key(),
			}
		}
	}()

	// Close the remove channel when both goroutines have completed
	go func() {
		latch.Wait()
		close(removeCh)
	}()

	ctx.JSON(http.StatusAccepted, gin.H{
		"success": true,
	})
}
