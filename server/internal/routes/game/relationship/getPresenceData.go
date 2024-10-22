package relationship

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"net/http"
)

func GetPresenceData(w http.ResponseWriter, r *http.Request) {
	middleware.Age2Game(r).Resources().ReturnSignedAsset("presenceData.json", &w, r, false)
}
