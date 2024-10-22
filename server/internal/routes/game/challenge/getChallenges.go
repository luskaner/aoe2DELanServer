package challenge

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"net/http"
)

func GetChallenges(w http.ResponseWriter, r *http.Request) {
	middleware.Age2Game(r).Resources().ReturnSignedAsset("challenges.json", &w, r, false)
}
