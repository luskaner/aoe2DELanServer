package advertisement

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/advertisement/shared"
	"net/http"
)

func Update(w http.ResponseWriter, r *http.Request) {
	var q shared.AdvertisementUpdateRequest
	if err := i.Bind(r, &q); err != nil {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	advertisements := middleware.Age2Game(r).Advertisements()
	adv, ok := advertisements.GetAdvertisement(q.Id)
	if !ok {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	advertisements.Update(adv, &q)
	i.JSON(&w,
		i.A{
			0,
			adv.Encode(),
		},
	)
}
