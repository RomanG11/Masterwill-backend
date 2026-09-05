package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const maxUploadBytes = 5 << 20 // 5 MiB — plenty for a product photo, small enough to not need chunking

var allowedPhotoTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// adminUploadPhoto stores a product photo under cfg.UploadsDir and returns
// the public URL it's served at (see the GET /uploads/ route in router.go).
// Validation is by sniffed content, not the client-supplied filename/header,
// since either can lie about what the bytes actually are.
func (a *api) adminUploadPhoto(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeError(w, http.StatusBadRequest, "файл завеликий або запит невалідний")
		return
	}

	file, _, err := r.FormFile("photo")
	if err != nil {
		writeError(w, http.StatusBadRequest, "додайте файл у полі photo")
		return
	}
	defer file.Close()

	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && err != io.ErrUnexpectedEOF {
		writeError(w, http.StatusBadRequest, "не вдалося прочитати файл")
		return
	}
	head = head[:n]

	contentType := http.DetectContentType(head)
	ext, ok := allowedPhotoTypes[contentType]
	if !ok {
		writeError(w, http.StatusBadRequest, "дозволені тільки JPEG, PNG або WEBP")
		return
	}

	name, err := randomFilename(ext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не вдалося створити файл")
		return
	}

	if err := os.MkdirAll(a.cfg.UploadsDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "не вдалося підготувати сховище")
		return
	}

	dst, err := os.Create(filepath.Join(a.cfg.UploadsDir, name))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не вдалося зберегти файл")
		return
	}
	defer dst.Close()

	if _, err := dst.Write(head); err != nil {
		writeError(w, http.StatusInternalServerError, "не вдалося зберегти файл")
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		writeError(w, http.StatusInternalServerError, "не вдалося зберегти файл")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"url": "/uploads/" + name})
}

func randomFilename(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate filename: %w", err)
	}
	return hex.EncodeToString(buf) + ext, nil
}
