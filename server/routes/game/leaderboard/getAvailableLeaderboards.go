package leaderboard

import (
	"github.com/luskaner/aoe2DELanServer/server/files"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"net/http"
)

func GetAvailableLeaderboards(w http.ResponseWriter, _ *http.Request) {
	i.JSON(&w, files.ArrayFiles["leaderboards.json"])
}
