package item

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"net/http"
)

func GetItemBundleItemsJson(w http.ResponseWriter, r *http.Request) {
	middleware.Age2Game(r).Resources().ReturnSignedAsset("itemBundleItems.json", &w, r, true)
}
