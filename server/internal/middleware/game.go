package middleware

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/server/internal/models/age2"
	"github.com/luskaner/aoe2DELanServer/server/internal/models/initializer"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

var gamePathHandlers = map[string]map[string]http.Handler{}

var ignoredPaths = map[string]bool{
	"/":                            true,
	"/test":                        true,
	"/game/msstore/getStoreTokens": true,
	"/wss/":                        true,
	"/game/news/getNews":           true,
}

func Age2Game(r *http.Request) age2.GameType {
	return Game[age2.GameType](r)
}

func Game[T any](r *http.Request) T {
	return r.Context().Value("game").(T)
}

func GameMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !ignoredPaths[r.URL.Path] && !strings.HasPrefix(r.URL.Path, "/cloudfiles/") {
			gameId := r.URL.Query().Get("title")
			if gameId == "" && r.Method == http.MethodPost {
				if err := r.ParseForm(); err == nil {
					gameId = r.Form.Get("title")
				}
			}
			if gameId == "" {
				session, ok := Session(r)
				if ok {
					gameId = session.GetGameId()
				}
			}
			gameSet := mapset.NewSet[string](viper.GetStringSlice("default.Games")...)
			if !gameSet.ContainsOne(gameId) {
				http.Error(w, "Unavailable game type", http.StatusBadRequest)
				return
			}
			ctx := context.WithValue(r.Context(), "game", initializer.Games[gameId])
			req := r.WithContext(ctx)
			if pathsHandlers, ok := gamePathHandlers[gameId]; ok {
				var handler http.Handler
				if handler, ok = pathsHandlers[r.Method]; ok {
					handler.ServeHTTP(w, req)
					return
				}
			}
			next.ServeHTTP(w, req)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
