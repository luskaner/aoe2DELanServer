package cloud

import (
	"fmt"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func GetTempCredentials(w http.ResponseWriter, r *http.Request) {
	fullKey := r.URL.Query().Get("key")
	key := strings.TrimPrefix(fullKey, "/cloudfiles/")
	cloudfiles := middleware.Age2Game(r).Resources().CloudFiles
	cred := cloudfiles.Credentials.CreateCredentials(key)
	t := cred.GetExpiry()
	tUnix := t.Unix()
	for _, file := range cloudfiles.Value {
		if file.Key == key {
			se := url.QueryEscape(t.Format(time.RFC3339))
			sv := url.QueryEscape(file.Version)
			i.JSON(&w, i.A{0, tUnix, fmt.Sprintf("sig=%s&se=%s&sv=%s&sp=r&sr=b", url.QueryEscape(cred.GetSignature()), se, sv), fullKey})
			return
		}
	}
	i.JSON(&w, i.A{2, t, "", fullKey})
}
