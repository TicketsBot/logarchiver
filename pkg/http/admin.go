package http

import (
	"errors"
	"github.com/TicketsBot/logarchiver/pkg/repository"
	"github.com/TicketsBot/logarchiver/pkg/repository/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
)

func (s *Server) adminListBuckets(ctx *gin.Context) {
	var buckets []model.Bucket
	if err := s.store.Tx(ctx, func(r repository.Repositories) (err error) {
		buckets, err = r.Buckets().ListBuckets(ctx)
		return
	}); err != nil {
		s.Logger.Error("Error fetching buckets", zap.Error(err))
		ctx.JSON(500, gin.H{
			"message": "Error fetching buckets",
		})
		return
	}

	ctx.JSON(200, buckets)
}

func (s *Server) adminCreateBucket(ctx *gin.Context) {
	type body struct {
		EndpointUrl string `json:"endpoint_url"`
		Name        string `json:"name"`
	}

	var b body
	if err := ctx.BindJSON(&b); err != nil {
		ctx.JSON(400, gin.H{
			"message": "missing endpoint_url or name",
		})
		return
	}

	var bucketId uuid.UUID
	if err := s.store.Tx(ctx, func(r repository.Repositories) (err error) {
		bucketId, err = r.Buckets().CreateBucket(ctx, b.EndpointUrl, b.Name)
		return
	}); err != nil {
		s.Logger.Error("Error creating bucket", zap.Error(err), zap.Any("request_data", b))
		ctx.JSON(500, gin.H{
			"message": "Error creating bucket",
		})
		return
	}

	ctx.JSON(http.StatusCreated, model.Bucket{
		Id:          bucketId,
		EndpointUrl: b.EndpointUrl,
		Name:        b.Name,
		Active:      false,
	})
}

func (s *Server) adminSetActiveBucket(ctx *gin.Context) {
	type body struct {
		BucketId uuid.UUID `json:"bucket_id"`
	}

	var b body
	if err := ctx.BindJSON(&b); err != nil {
		ctx.JSON(400, gin.H{
			"message": "missing bucket_id",
		})
		return
	}

	var ErrBucketNotFound = errors.New("bucket not found")
	if err := s.store.Tx(ctx, func(r repository.Repositories) error {
		buckets, err := r.Buckets().ListBuckets(ctx)
		if err != nil {
			return err
		}

		found := false
		for _, bucket := range buckets {
			if bucket.Id == b.BucketId {
				found = true
				break
			}
		}

		if !found {
			return ErrBucketNotFound
		}

		return r.Buckets().SetActiveBucket(ctx, b.BucketId)
	}); err != nil {
		if errors.Is(err, ErrBucketNotFound) {
			ctx.JSON(404, gin.H{
				"message": "bucket not found",
			})
		} else {
			s.Logger.Error("Error setting active bucket", zap.Error(err), zap.Any("request_data", b))
			ctx.JSON(500, gin.H{
				"message": "Error setting active bucket",
			})
		}

		return
	}

	ctx.Status(204)
}
