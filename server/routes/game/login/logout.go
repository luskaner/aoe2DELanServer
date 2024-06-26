package login

import (
	"net/http"
	i "server/internal"
	"server/middleware"
	"server/models"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := middleware.Session(r)
	u := sess.GetUser()
	advs := models.FindAdvertisements(func(adv *models.Advertisement) bool {
		_, found := adv.GetPeer(u)
		return found
	})
	for _, adv := range advs {
		adv.RemovePeer(u)
	}
	u.SetPresence(0)
	sess.Delete()
	i.JSON(&w, i.A{0})
}
