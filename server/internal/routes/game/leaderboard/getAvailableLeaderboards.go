package leaderboard

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"net/http"
)

func GetAvailableLeaderboards(w http.ResponseWriter, r *http.Request) {
	i.JSON(&w, middleware.Age2Game(r).Resources().ArrayFiles["leaderboards.json"])
}
