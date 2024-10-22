package age2

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/advertisement/shared"
)

type Game struct {
	resources      *models.MainResources
	users          *models.MainUsers
	advertisements *models.MainAdvertisements
}

type GameType = models.Game[
	*models.MainUser,
	*models.MainUsers,
	*models.MainPeer,
	*models.MainMessage,
	*models.MainAdvertisement,
	*shared.AdvertisementHostRequest,
	*shared.AdvertisementUpdateRequest,
	*models.MainAdvertisements,
]

func CreateAge2Game() GameType {
	game := &Game{
		resources:      &models.MainResources{},
		users:          &models.MainUsers{},
		advertisements: &models.MainAdvertisements{},
	}
	keyedFilenames := mapset.NewSet[string]("itemBundleItems.json", "itemDefinitions.json")
	game.resources.Initialize(common.GameAoE2, keyedFilenames)
	game.users.Initialize()
	game.advertisements.Initialize(game.users)
	return game
}

func (g *Game) Resources() *models.MainResources {
	return g.resources
}

func (g *Game) Users() *models.MainUsers {
	return g.users
}

func (g *Game) Advertisements() *models.MainAdvertisements {
	return g.advertisements
}
