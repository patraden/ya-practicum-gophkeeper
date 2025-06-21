package mockminio

// import (
// 	"io"
// 	"net/http"
// 	"sync"

// 	"github.com/go-chi/chi/v5"
// )

// // S3MockHandler mocks S3/MinIO object storage HTTP API.
// type S3MockHandler struct {
// 	mu      sync.RWMutex
// 	storage map[string][]byte // map[objectName]content
// }

// func NewS3MockHandler() *S3MockHandler {
// 	return &S3MockHandler{
// 		storage: make(map[string][]byte),
// 	}
// }

// // RegisterRoutes registers S3 mock endpoints to the router.
// func (h *S3MockHandler) RegisterRoutes(router chi.Router) {
// 	router.Get("/mock-bucket/{objectName}", h.GetObject)
// 	router.Put("/mock-bucket/{objectName}", h.PutObject)
// 	// You can add DeleteObject, ListObjects, etc here as needed.
// }

// // GetObject mocks retrieving an object from the bucket.
// func (h *S3MockHandler) GetObject(w http.ResponseWriter, r *http.Request) {
// 	objectName := chi.URLParam(r, "objectName")

// 	h.mu.RLock()
// 	content, exists := h.storage[objectName]
// 	h.mu.RUnlock()

// 	if !exists {
// 		http.Error(w, "object not found", http.StatusNotFound)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	w.Write(content)
// }

// // PutObject mocks storing an object in the bucket.
// func (h *S3MockHandler) PutObject(w http.ResponseWriter, r *http.Request) {
// 	objectName := chi.URLParam(r, "objectName")

// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		http.Error(w, "failed to read body", http.StatusInternalServerError)
// 		return
// 	}
// 	defer r.Body.Close()

// 	h.mu.Lock()
// 	h.storage[objectName] = body
// 	h.mu.Unlock()

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("stored object " + objectName))
// }
