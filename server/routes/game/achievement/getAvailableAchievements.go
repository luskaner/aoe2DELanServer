package achievement

import (
	"github.com/luskaner/aoe2DELanServer/server/files"
	"net/http"
)

func GetAvailableAchievements(w http.ResponseWriter, r *http.Request) {
	files.ReturnSignedAsset("achievements.json", &w, r, false)
}
