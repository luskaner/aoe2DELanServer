package leaderboard

import (
	"net/http"
	i "server/internal"
)

func ApplyOfflineUpdates(w http.ResponseWriter, _ *http.Request) {
	// TODO: Implement? which kind of updates?
	i.JSON(&w, i.A{0, i.A{}, i.A{}})
}
