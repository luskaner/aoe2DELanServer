package relationship

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/files"
	"net/http"
)

func GetPresenceData(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("presenceData.json", &w, r, false)
}
