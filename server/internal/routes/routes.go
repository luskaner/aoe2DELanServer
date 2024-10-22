package routes

import (
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/cloudfiles"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/account"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/achievement"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/advertisement"
	Automatch2 "github.com/luskaner/aoe2DELanServer/server/internal/routes/game/automatch2"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/challenge"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/chat"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/clan"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/cloud"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/communityEvent"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/invitation"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/item"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/leaderboard"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/login"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/news"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/party"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/relationship"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/msstore"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/test"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/wss"
	"net/http"
)

type Group struct {
	parent *Group
	path   string
	mux    *http.ServeMux
}

func (g *Group) fullPath() string {
	if g.parent == nil {
		return g.path
	}
	return g.parent.fullPath() + g.path
}

func (g *Group) Subgroup(path string) *Group {
	return &Group{
		parent: g,
		path:   path,
		mux:    g.mux,
	}
}

func (g *Group) HandleFunc(method string, path string, handler http.HandlerFunc) {
	g.mux.HandleFunc(method+" "+g.fullPath()+path, handler)
}

func Initialize(mux *http.ServeMux) {
	baseGroup := Group{
		path: "",
		mux:  mux,
	}
	gameGroup := baseGroup.Subgroup("/game")
	itemGroup := gameGroup.Subgroup("/item")
	itemGroup.HandleFunc("GET", "/getItemBundleItemsJson", item.GetItemBundleItemsJson)
	itemGroup.HandleFunc("GET", "/getItemDefinitionsJson", item.GetItemDefinitionsJson)
	itemGroup.HandleFunc("GET", "/getItemLoadouts", item.GetItemLoadouts)
	itemGroup.HandleFunc("POST", "/signItems", item.SignItems)
	itemGroup.HandleFunc("GET", "/getInventoryByProfileIDs", item.GetInventoryByProfileIDs)

	clanGroup := gameGroup.Subgroup("/clan")
	clanGroup.HandleFunc("POST", "/create", clan.Create)
	clanGroup.HandleFunc("GET", "/find", clan.Find)

	communityEventGroup := gameGroup.Subgroup("/CommunityEvent")
	communityEventGroup.HandleFunc("GET", "/getAvailableCommunityEvents", communityEvent.GetAvailableCommunityEvents)

	challengeGroup := gameGroup.Subgroup("/Challenge")
	challengeGroup.HandleFunc("GET", "/getChallengeProgress", challenge.GetChallengeProgress)
	challengeGroup.HandleFunc("GET", "/getChallenges", challenge.GetChallenges)

	newsGroup := gameGroup.Subgroup("/news")
	newsGroup.HandleFunc("GET", "/getNews", news.GetNews)

	loginGroup := gameGroup.Subgroup("/login")
	loginGroup.HandleFunc("POST", "/platformlogin", login.Platformlogin)
	loginGroup.HandleFunc("POST", "/logout", login.Logout)

	accountGroup := gameGroup.Subgroup("/account")
	accountGroup.HandleFunc("POST", "/setLanguage", account.SetLanguage)
	accountGroup.HandleFunc("POST", "/setCrossplayEnabled", account.SetCrossplayEnabled)
	accountGroup.HandleFunc("POST", "/setAvatarMetadata", account.SetAvatarMetadata)
	accountGroup.HandleFunc("POST", "/FindProfilesByPlatformID", account.FindProfilesByPlatformID)
	accountGroup.HandleFunc("GET", "/FindProfiles", account.FindProfiles)
	accountGroup.HandleFunc("GET", "/getProfileName", account.GetProfileName)

	LeaderboardGroup := gameGroup.Subgroup("/Leaderboard")
	LeaderboardGroup.HandleFunc("GET", "/getRecentMatchHistory", leaderboard.GetRecentMatchHistory)
	LeaderboardGroup.HandleFunc("GET", "/getLeaderBoard", leaderboard.GetLeaderBoard)
	LeaderboardGroup.HandleFunc("GET", "/getAvailableLeaderboards", leaderboard.GetAvailableLeaderboards)
	LeaderboardGroup.HandleFunc("GET", "/getStatGroupsByProfileIDs", leaderboard.GetStatGroupsByProfileIDs)
	LeaderboardGroup.HandleFunc("GET", "/getStatsForLeaderboardByProfileName", leaderboard.GetStatsForLeaderboardByProfileName)
	LeaderboardGroup.HandleFunc("GET", "/getPartyStat", leaderboard.GetPartyStat)

	leaderboardGroup := gameGroup.Subgroup("/leaderboard")
	leaderboardGroup.HandleFunc("POST", "/applyOfflineUpdates", leaderboard.ApplyOfflineUpdates)
	leaderboardGroup.HandleFunc("POST", "/setAvatarStatValues", leaderboard.SetAvatarStatValues)

	automatch2Group := gameGroup.Subgroup("/automatch2")
	automatch2Group.HandleFunc("GET", "/getAutomatchMap", Automatch2.GetAutomatchMap)

	AchievementGroup := gameGroup.Subgroup("/Achievement")
	AchievementGroup.HandleFunc("GET", "/getAchievements", achievement.GetAchievements)
	AchievementGroup.HandleFunc("GET", "/getAvailableAchievements", achievement.GetAvailableAchievements)

	achievementGroup := gameGroup.Subgroup("/achievement")
	achievementGroup.HandleFunc("POST", "/applyOfflineUpdates", achievement.ApplyOfflineUpdates)
	achievementGroup.HandleFunc("POST", "/grantAchievement", achievement.GrantAchievement)
	achievementGroup.HandleFunc("POST", "/syncStats", achievement.SyncStats)

	advertisementGroup := gameGroup.Subgroup("/advertisement")
	advertisementGroup.HandleFunc("POST", "/updatePlatformSessionID", advertisement.UpdatePlatformSessionID)
	advertisementGroup.HandleFunc("POST", "/join", advertisement.Join)
	advertisementGroup.HandleFunc("POST", "/updateTags", advertisement.UpdateTags)
	advertisementGroup.HandleFunc("POST", "/update", advertisement.Update)
	advertisementGroup.HandleFunc("POST", "/leave", advertisement.Leave)
	advertisementGroup.HandleFunc("POST", "/host", advertisement.Host)
	advertisementGroup.HandleFunc("GET", "/getLanAdvertisements", advertisement.GetLanAdvertisements)
	advertisementGroup.HandleFunc("GET", "/findObservableAdvertisements", advertisement.FindObservableAdvertisements)
	advertisementGroup.HandleFunc("GET", "/getAdvertisements", advertisement.GetAdvertisements)
	advertisementGroup.HandleFunc("GET", "/findAdvertisements", advertisement.FindAdvertisements)
	advertisementGroup.HandleFunc("POST", "/updateState", advertisement.UpdateState)

	chatGroup := gameGroup.Subgroup("/chat")
	chatGroup.HandleFunc("GET", "/getChatChannels", chat.GetChatChannels)
	chatGroup.HandleFunc("GET", "/getOfflineMessages", chat.GetOfflineMessages)

	relationshipGroup := gameGroup.Subgroup("/relationship")
	relationshipGroup.HandleFunc("GET", "/getRelationships", relationship.GetRelationships)
	relationshipGroup.HandleFunc("GET", "/getPresenceData", relationship.GetPresenceData)
	relationshipGroup.HandleFunc("POST", "/setPresence", relationship.SetPresence)
	relationshipGroup.HandleFunc("POST", "/ignore", relationship.Ignore)
	relationshipGroup.HandleFunc("POST", "/clearRelationship", relationship.ClearRelationship)

	partyGroup := gameGroup.Subgroup("/party")
	partyGroup.HandleFunc("POST", "/peerAdd", party.PeerAdd)
	partyGroup.HandleFunc("POST", "/peerUpdate", party.PeerUpdate)
	partyGroup.HandleFunc("POST", "/sendMatchChat", party.SendMatchChat)
	partyGroup.HandleFunc("POST", "/reportMatch", party.ReportMatch)
	partyGroup.HandleFunc("POST", "/finalizeReplayUpload", party.FinalizeReplayUpload)
	partyGroup.HandleFunc("POST", "/updateHost", party.UpdateHost)

	invitationGroup := gameGroup.Subgroup("/invitation")
	invitationGroup.HandleFunc("POST", "/extendInvitation", invitation.ExtendInvitation)
	invitationGroup.HandleFunc("POST", "/cancelInvitation", invitation.CancelInvitation)
	invitationGroup.HandleFunc("POST", "/replyToInvitation", invitation.ReplyToInvitation)

	cloudGroup := gameGroup.Subgroup("/cloud")
	cloudGroup.HandleFunc("GET", "/getFileURL", cloud.GetFileURL)
	cloudGroup.HandleFunc("GET", "/getTempCredentials", cloud.GetTempCredentials)

	msstoreGroup := gameGroup.Subgroup("/msstore")
	msstoreGroup.HandleFunc("GET", "/getStoreTokens", msstore.GetStoreTokens)

	// Used for the launcher
	baseGroup.HandleFunc("GET", "/test", test.Test)
	baseGroup.HandleFunc("GET", "/wss/", wss.Handle)
	baseGroup.HandleFunc("GET", "/cloudfiles/", cloudfiles.Cloudfiles)
}
