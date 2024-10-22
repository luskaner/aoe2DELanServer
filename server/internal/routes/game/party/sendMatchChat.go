package party

import (
	"encoding/json"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/wss"
	"net/http"
)

type request struct {
	ToProfileIdsStr string `schema:"to_profile_ids"`
	MessageTypeID   uint8  `schema:"messageTypeID"`
	MatchID         int32  `schema:"match_id"`
	Broadcast       bool   `schema:"broadcast"`
	Message         string `schema:"message"`
}

func SendMatchChat(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := i.Bind(r, &req); err != nil {
		i.JSON(&w, i.A{2})
		return
	}

	var toProfileIds []int32
	err := json.Unmarshal([]byte(req.ToProfileIdsStr), &toProfileIds)
	if err != nil {
		i.JSON(&w, i.A{2})
		return
	}
	game := middleware.Age2Game(r)
	adv, ok := game.Advertisements().GetAdvertisement(req.MatchID)
	if !ok {
		i.JSON(&w, i.A{2})
		return
	}

	sess, _ := middleware.Session(r)
	currentUser, _ := game.Users().GetUserById(sess.GetUserId())

	// Only peers within the match can send messages
	// What about AI?
	if _, ok = adv.GetPeer(currentUser); !ok {
		i.JSON(&w, i.A{2})
		return
	}

	users := game.Users()

	receivers := make([]*models.MainUser, len(toProfileIds))
	for j, profileId := range toProfileIds {
		receivers[j], _ = users.GetUserById(profileId)
	}

	message := adv.AddMessage(
		req.Broadcast,
		req.Message,
		req.MessageTypeID,
		currentUser,
		receivers,
	)

	messageEncoded := message.Encode()
	var receiverSession *models.Session
	for _, receiver := range receivers {
		receiverSession, ok = models.GetSessionByUserId(receiver.GetId())
		if !ok {
			continue
		}
		go func(receiverSession string, receiverUserId int32, messageEncoded i.A) {
			wss.SendMessage(
				receiverSession,
				i.A{
					0,
					"MatchReceivedChatMessage",
					receiverUserId,
					messageEncoded,
				},
			)
		}(receiverSession.GetId(), receiver.GetId(), messageEncoded)
	}
	i.JSON(&w, i.A{0})
}
