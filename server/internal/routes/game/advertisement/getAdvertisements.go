package advertisement

import (
	"encoding/json"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"net/http"
)

func GetAdvertisements(w http.ResponseWriter, r *http.Request) {
	matchIdsStr := r.URL.Query().Get("match_ids")
	var advsIds []int32
	err := json.Unmarshal([]byte(matchIdsStr), &advsIds)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	advs := middleware.Age2Game(r).Advertisements().FindAdvertisementsEncoded(func(adv *models.MainAdvertisement) bool {
		for _, advId := range advsIds {
			if adv.GetId() == advId {
				return true
			}
		}
		return false
	})
	if advs == nil {
		i.JSON(&w,
			i.A{0, i.A{}},
		)
	} else {
		i.JSON(&w,
			i.A{0, advs},
		)
	}
}
