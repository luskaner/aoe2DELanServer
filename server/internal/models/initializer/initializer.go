package initializer

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/server/internal/models/age2"
)

var Games = map[string]any{}

func InitializeGames(gameIds mapset.Set[string]) {
	for gameId := range gameIds.Iter() {
		var game any
		switch gameId {
		case common.GameAoE2:
			game = age2.CreateAge2Game()
		}
		Games[gameId] = game
	}
}
