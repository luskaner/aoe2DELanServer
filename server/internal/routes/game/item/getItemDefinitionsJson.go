package item

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"net/http"
)

func GetItemDefinitionsJson(w http.ResponseWriter, r *http.Request) {
	middleware.Age2Game(r).Resources().ReturnSignedAsset("itemDefinitions.json", &w, r, true)
}
