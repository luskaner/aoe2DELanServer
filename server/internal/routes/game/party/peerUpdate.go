package party

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/party/shared"
	"net/http"
)

func PeerUpdate(w http.ResponseWriter, r *http.Request) {
	// What about isNonParticipants[]? observers? ai players?
	adv, length, profileIds, raceIds, statGroupIds, teamIds := shared.ParseParameters(r)
	if adv == nil {
		i.JSON(&w, i.A{2})
		return
	}
	sess, _ := middleware.Session(r)
	users := make([]*models.MainUser, length)
	gameUsers := middleware.Age2Game(r).Users()
	currentUser, _ := gameUsers.GetUserById(sess.GetUserId())
	// Only the host can update peers
	if adv.GetHost() != currentUser {
		i.JSON(&w, i.A{2})
		return
	}

	for j := 0; j < length; j++ {
		u, ok := gameUsers.GetUserById(profileIds[j])
		if !ok || u.GetStatId() != statGroupIds[j] {
			i.JSON(&w, i.A{2})
			return
		}
		users[j] = u
	}
	for j, u := range users {
		adv.UpdatePeer(u, raceIds[j], teamIds[j])
	}
	i.JSON(&w, i.A{0})
}
