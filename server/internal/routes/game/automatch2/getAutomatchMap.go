package Automatch2

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"net/http"
)

func GetAutomatchMap(w http.ResponseWriter, r *http.Request) {
	i.JSON(&w, middleware.Age2Game(r).Resources().ArrayFiles["automatchMaps.json"])
}
