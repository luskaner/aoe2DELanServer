package userData

import "github.com/luskaner/aoe2DELanServer/common"

func Metadata(gameId string) Data {
	var path string
	switch gameId {
	case common.GameAoE2:
		path = "metadata"
	}
	return Data{Path: path}
}
