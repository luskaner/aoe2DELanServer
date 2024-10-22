package party

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/party/shared"
	"net/http"
)

func PeerAdd(w http.ResponseWriter, r *http.Request) {
	adv, length, profileIds, raceIds, statGroupIds, teamIds := shared.ParseParameters(r)
	if adv == nil {
		i.JSON(&w, i.A{2})
		return
	}
	sess, _ := middleware.Session(r)
	game := middleware.Age2Game(r)
	gameUsers := game.Users()
	currentUser, _ := gameUsers.GetUserById(sess.GetUserId())
	// Only the host can add peers
	host := adv.GetHost()
	if host != currentUser {
		i.JSON(&w, i.A{2})
		return
	}
	users := make([]*models.MainUser, length)
	for j := 0; j < length; j++ {
		u, ok := gameUsers.GetUserById(profileIds[j])
		if !ok || u.GetStatId() != statGroupIds[j] {
			i.JSON(&w, i.A{2})
			return
		}
		users[j] = u
	}
	advertisements := game.Advertisements()
	for j, u := range users {
		advertisements.NewPeer(adv, u, raceIds[j], teamIds[j])
	}
	i.JSON(&w, i.A{0})
}
