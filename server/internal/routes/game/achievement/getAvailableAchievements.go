package achievement

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"net/http"
)

func GetAvailableAchievements(w http.ResponseWriter, r *http.Request) {
	middleware.Age2Game(r).Resources().ReturnSignedAsset("achievements.json", &w, r, false)
}
