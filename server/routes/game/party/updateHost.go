package party

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/models"
	"net/http"
	"strconv"
)

func UpdateHost(w http.ResponseWriter, r *http.Request) {
	advStr := r.PostFormValue("match_id")
	advId, err := strconv.ParseInt(advStr, 10, 32)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	adv, ok := models.GetAdvertisement(int32(advId))
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}
	ok = adv.UpdateHost()
	var code int
	if ok {
		code = 1
	} else {
		code = 2
	}
	i.JSON(&w, i.A{code})
}
