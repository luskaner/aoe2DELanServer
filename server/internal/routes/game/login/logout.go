package login

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := middleware.Session(r)
	game := middleware.Age2Game(r)
	u, _ := game.Users().GetUserById(sess.GetUserId())
	advertisements := middleware.Age2Game(r).Advertisements()
	advIds := advertisements.FindAdvertisements(func(adv *models.MainAdvertisement) bool {
		_, found := adv.GetPeer(u)
		return found
	})
	for _, advId := range advIds {
		advertisements.RemovePeer(advId, u)
	}
	u.SetPresence(0)
	sess.Delete()
	i.JSON(&w, i.A{0})
}
