package models

import (
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type User interface {
	GetId() int32
	GetStatId() int32
	GetProfileId() int32
	GetReliclink() int32
	GetAlias() string
	GetPlatformId() int
	GetPlatformPath() string
	GetPlatformUserID() uint64
	GetExtraProfileInfo() i.A
	SetPresence(value int8)
	GetPresence() int8
	GetProfileInfo(includePresence bool) i.A
}

type ComparableUser interface {
	User
	comparable
}

type Users[U User] interface {
	GetOrCreateUser(remoteAddr string, isXbox bool, platformUserId uint64, alias string) U
	GetUserByStatId(id int32) (U, bool)
	GetUserById(id int32) (U, bool)
	GetProfileInfo(includePresence bool, matches func(user U) bool) []i.A
}

type Peer[U User] interface {
	GetAdvertisementId() int32
	GetUser() U
	GetRace() int32
	GetTeam() int32
	Encode() i.A
	Invite(invitee U)
	Uninvite(invitee U)
	IsInvited(invitee U) bool
	Update(race int32, team int32)
}

type Message[U User] interface {
	GetTime() int64
	GetBroadcast() bool
	GetContent() string
	GetType() uint8
	GetSender() U
	GetReceivers() []U
	GetAdvertisementId() int32
	Encode() i.A
}

type Advertisement[U ComparableUser, P Peer[U], M Message[U]] interface {
	GetModDllChecksum() uint32
	GetModDllFile() string
	GetPasswordValue() string
	GetStartTime() int64
	GetState() int8
	GetId() int32
	GetDescription() string
	GetRelayRegion() string
	GetJoinable() bool
	GetVisible() bool
	GetHost() U
	GetAppBinaryChecksum() uint32
	GetDataChecksum() uint32
	GetMatchType() uint8
	GetModName() string
	GetModVersion() string
	GetIp() string
	GetVersionFlags() uint32
	GetPeers() *orderedmap.OrderedMap[U, P]
	GetPeer(user U) (P, bool)
	AddMessage(broadcast bool, content string, typeId uint8, sender U, receivers []U) M
	UpdatePeer(user U, race int32, team int32)
	UpdateState(state int8)
	Encode() i.A
	EncodePeers() i.A
}

type Advertisements[U ComparableUser, P Peer[U], M Message[U], A Advertisement[U, P, M], D any, S any] interface {
	GetAdvertisement(id int32) (A, bool)
	Store(advFrom S) A
	Update(adv A, advFrom D)
	Delete(adv A)
	NewPeer(adv A, u U, race int32, team int32) P
	RemovePeer(adv A, user U)
	FindAdvertisements(matches func(adv A) bool) []A
	FindAdvertisementsEncoded(matches func(adv A) bool) []i.A
	IsInAdvertisement(user U) bool
	IsPeer(user U) bool
	IsHost(user U) bool
}

type Game[
	U ComparableUser,
	R Users[U],
	P Peer[U],
	M Message[U],
	A Advertisement[U, P, M],
	D any,
	S any,
	V Advertisements[U, P, M, A, S, D],
] interface {
	Resources() *MainResources
	Users() R
	Advertisements() V
}
