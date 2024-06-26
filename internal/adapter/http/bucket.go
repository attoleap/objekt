package http // import "github.com/attoleap/objekt/internal/adapter/http"

import (
	"encoding/json"
	"net/http"

	"github.com/attoleap/objekt/internal/core/domain"
	"github.com/attoleap/objekt/internal/core/port"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
)

type BucketHandler struct {
	log    *zerolog.Logger
	router *httprouter.Router
	svc    port.BucketService
}

func NewBucketHandler(log *zerolog.Logger, router *httprouter.Router, svc port.BucketService) *BucketHandler {
	return &BucketHandler{
		log:    log,
		router: router,
		svc:    svc,
	}
}

type createBucketRequest struct {
	Name   string              `json:"name"`
	Type   domain.BucketType   `json:"type"`
	Region domain.BucketRegion `json:"region"`
}

func (h *BucketHandler) AddRoutes() {
	h.router.GET("/buckets", h.ListBuckets)
	h.router.POST("/buckets", contentTypeMiddleware(h.CreateBucket, []string{ContentTypeJSON}))
	h.router.GET("/buckets/:bucket_id", h.GetBucket)
	h.router.DELETE("/buckets/:bucket_id", h.DeleteBucket)
}

func (h *BucketHandler) CreateBucket(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var requestBody createBucketRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		h.log.Err(err).Msg("failed to decode request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bucket := &domain.Bucket{
		Name:   requestBody.Name,
		Type:   requestBody.Type,
		Region: requestBody.Region,
	}

	bucket, err = h.svc.CreateBucket(r.Context(), bucket)
	if err != nil {
		h.log.Err(err).Str("bucket_name", requestBody.Name).Msg("failed to create bucket")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.Header().Set(HeaderLocation, "/buckets/"+bucket.ID.String())
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(bucket); err != nil {
		h.log.Err(err).Msg("failed to encode CreateBucket response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *BucketHandler) GetBucket(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	bucketID := p.ByName("bucket_id")
	bucket, err := h.svc.GetBucket(r.Context(), bucketID)
	if err != nil {
		h.log.Err(err).Str("bucket_id", bucketID).Msg("failed to retrieve bucket")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	if err = json.NewEncoder(w).Encode(bucket); err != nil {
		h.log.Err(err).Msg("failed to encode GetBucket response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *BucketHandler) ListBuckets(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	buckets, err := h.svc.ListBuckets(r.Context())
	if err != nil {
		h.log.Err(err).Msg("failed to list buckets")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	if err = json.NewEncoder(w).Encode(buckets); err != nil {
		h.log.Err(err).Msg("failed to encode ListBuckets response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *BucketHandler) DeleteBucket(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	bucketID := p.ByName("bucket_id")
	err := h.svc.DeleteBucket(r.Context(), bucketID)
	if err != nil {
		h.log.Err(err).Str("bucket_id", bucketID).Msg("failed to delete bucket")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
