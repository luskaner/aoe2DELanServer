package leaderboard

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/leaderboard/shared"
	"net/http"
)

func GetPartyStat(w http.ResponseWriter, r *http.Request) {
	response := shared.GetStatGroups(r, r.URL.Query().Get("statsids"), false, true)
	i.JSON(&w, response)
}
