package cloud

import (
	"encoding/json"
	"fmt"
	"github.com/luskaner/aoe2DELanServer/common"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"net/http"
)

func GetFileURL(w http.ResponseWriter, r *http.Request) {
	namesStr := r.URL.Query().Get("names")
	var names []string
	err := json.Unmarshal([]byte(namesStr), &names)
	if err != nil {
		i.JSON(&w, i.A{2, i.A{}})
		return
	}
	game := middleware.Age2Game(r)
	descriptions := make(i.A, len(names))
	for j, name := range names {
		fileData := game.Resources().CloudFiles.Value[name]
		finalPart := fileData.Key
		descriptions[j] = i.A{
			name,
			fileData.Length,
			fileData.Id,
			fmt.Sprintf("https://%s/cloudfiles/%s", common.Domain, finalPart),
			finalPart,
		}
	}
	i.JSON(&w, i.A{0, descriptions})
}
