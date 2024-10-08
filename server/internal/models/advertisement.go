package models

import (
	"fmt"
	i "github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes/game/advertisement/shared"
	"github.com/wk8/go-ordered-map/v2"
	"math"
	"sync"
	"time"
)

type ModDll struct {
	file     string
	checksum uint32
}

type Observers struct {
	enabled  bool
	delay    bool
	password string
}

type Password struct {
	value   string
	enabled bool
}

type Advertisement struct {
	id                int32
	ip                string
	automatchPollId   int32
	relayRegion       string
	appBinaryChecksum uint32
	mapName           string
	description       string
	dataChecksum      uint32
	host              *Peer
	modDll            ModDll
	modName           string
	modVersion        string
	observers         Observers
	password          Password
	visible           bool
	party             int32
	race              int32
	team              int32
	statGroup         int32
	versionFlags      uint32
	joinable          bool
	matchType         uint8
	maxPlayers        uint8
	options           string
	slotInfo          string
	platformSessionId uint64
	state             int8
	startTime         int64
	chat              []Message
	peers             *orderedmap.OrderedMap[*User, *Peer]
}

var peers = make(map[*User]interface{})
var hosts = make(map[*User]interface{})
var advertisementStore = make(map[int32]*Advertisement)
var chatLock = sync.RWMutex{}
var advLock = i.NewKeyRWMutex()
var peerLock = i.NewKeyRWMutex()
var hostLock = i.NewKeyRWMutex()

func (adv *Advertisement) GetModDllChecksum() uint32 {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.modDll.checksum
}

func (adv *Advertisement) GetModDllFile() string {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.modDll.file
}

func (adv *Advertisement) GetPasswordValue() string {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.password.value
}

func (adv *Advertisement) GetStartTime() int64 {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.startTime
}

func (adv *Advertisement) GetState() int8 {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.state
}

func (adv *Advertisement) GetId() int32 {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.id
}

func (adv *Advertisement) GetDescription() string {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.description
}

func (adv *Advertisement) GetRelayRegion() string {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.relayRegion
}

func (adv *Advertisement) GetJoinable() bool {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.joinable
}

func (adv *Advertisement) GetVisible() bool {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.visible
}

func (adv *Advertisement) GetHost() *Peer {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.host
}

func (adv *Advertisement) GetAppBinaryChecksum() uint32 {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.appBinaryChecksum
}

func (adv *Advertisement) GetDataChecksum() uint32 {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.dataChecksum
}

func (adv *Advertisement) GetMatchType() uint8 {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.matchType
}

func (adv *Advertisement) GetModName() string {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.modName
}

func (adv *Advertisement) GetModVersion() string {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.modVersion
}

func (adv *Advertisement) GetIp() string {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.ip
}

func (adv *Advertisement) GetVersionFlags() uint32 {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.versionFlags
}

func (adv *Advertisement) GetPeers() *orderedmap.OrderedMap[*User, *Peer] {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	return adv.peers
}

func (adv *Advertisement) GetPeer(user *User) (*Peer, bool) {
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	userId := user.GetId()
	peerLock.RLock(userId)
	defer peerLock.RUnlock(userId)
	u, exists := adv.peers.Get(user)
	if !exists {
		return nil, false
	}
	return u, true
}

func StoreAdvertisement(advFrom *shared.AdvertisementHostRequest) *Advertisement {
	if advFrom.Id != -1 {
		return nil
	}
	var id int32
	for {
		id = i.Rng.Int31n(math.MaxInt32)
		advLock.RLock(id)
		_, exists := advertisementStore[id]
		advLock.RUnlock(id)
		if !exists {
			adv := &Advertisement{}
			adv.id = id
			adv.ip = fmt.Sprintf("/10.0.11.%d", i.Rng.Intn(255)+1)
			adv.relayRegion = advFrom.RelayRegion
			adv.party = advFrom.Party
			adv.race = advFrom.Race
			adv.team = advFrom.Team
			adv.statGroup = advFrom.StatGroup
			adv.peers = orderedmap.New[*User, *Peer]()
			adv.chat = make([]Message, 0)
			u, _ := GetUserById(advFrom.HostId)
			adv.NewPeer(u, advFrom.Race, advFrom.Team)
			adv.Update(&shared.AdvertisementUpdateRequest{
				Id:                adv.id,
				AppBinaryChecksum: advFrom.AppBinaryChecksum,
				DataChecksum:      advFrom.DataChecksum,
				ModDllChecksum:    advFrom.ModDllChecksum,
				ModDllFile:        advFrom.ModDllFile,
				ModName:           advFrom.ModName,
				ModVersion:        advFrom.ModVersion,
				VersionFlags:      advFrom.VersionFlags,
				Description:       advFrom.Description,
				AutomatchPollId:   advFrom.AutomatchPollId,
				MapName:           advFrom.MapName,
				HostId:            advFrom.HostId,
				Observable:        advFrom.Observable,
				ObserverPassword:  advFrom.ObserverPassword,
				ObserverDelay:     advFrom.ObserverDelay,
				Password:          advFrom.Password,
				Passworded:        advFrom.Passworded,
				Visible:           advFrom.Visible,
				Joinable:          advFrom.Joinable,
				MatchType:         advFrom.MatchType,
				MaxPlayers:        advFrom.MaxPlayers,
				Options:           advFrom.Options,
				SlotInfo:          advFrom.SlotInfo,
				PlatformSessionId: advFrom.PlatformSessionId,
				State:             advFrom.State,
			})
			advLock.Lock(id)
			advertisementStore[id] = adv
			advLock.Unlock(id)
			return adv
		}
	}
}

func (adv *Advertisement) AddMessage(broadcast bool, content string, typeId uint8, sender *User, receivers []*User) *Message {
	message := &Message{
		advertisementId: adv.GetId(),
		time:            time.Now().UTC().Unix(),
		broadcast:       broadcast,
		content:         content,
		typ:             typeId,
		sender:          sender,
		receivers:       receivers,
	}
	chatLock.Lock()
	defer chatLock.Unlock()
	adv.chat = append(adv.chat, *message)
	return message
}

func (adv *Advertisement) Update(advFrom *shared.AdvertisementUpdateRequest) {
	advLock.Lock(adv.id)
	if adv.host != nil {
		previousHostId := adv.host.GetUser().GetId()
		hostLock.Lock(previousHostId)
		removeHost(adv.host)
		hostLock.Unlock(previousHostId)
	}
	u, _ := GetUserById(advFrom.HostId)
	adv.host, _ = adv.peers.Get(u)
	hostLock.Lock(advFrom.HostId)
	addHost(adv.host)
	hostLock.Unlock(advFrom.HostId)
	adv.automatchPollId = advFrom.AutomatchPollId
	adv.appBinaryChecksum = advFrom.AppBinaryChecksum
	adv.mapName = advFrom.MapName
	adv.description = advFrom.Description
	adv.dataChecksum = advFrom.DataChecksum
	adv.modDll.checksum = advFrom.ModDllChecksum
	adv.modDll.file = advFrom.ModDllFile
	adv.modName = advFrom.ModName
	adv.modVersion = advFrom.ModVersion
	adv.observers.delay = advFrom.ObserverDelay
	adv.observers.enabled = advFrom.Observable
	adv.observers.password = advFrom.ObserverPassword
	adv.password.enabled = advFrom.Passworded
	adv.password.value = advFrom.Password
	adv.visible = advFrom.Visible
	adv.versionFlags = advFrom.VersionFlags
	adv.joinable = advFrom.Joinable
	adv.matchType = advFrom.MatchType
	adv.maxPlayers = advFrom.MaxPlayers
	adv.options = advFrom.Options
	adv.slotInfo = advFrom.SlotInfo
	adv.platformSessionId = advFrom.PlatformSessionId
	advLock.Unlock(adv.id)
	adv.UpdateState(advFrom.State)
}

func GetAdvertisement(id int32) (*Advertisement, bool) {
	advLock.RLock(id)
	defer advLock.RUnlock(id)
	adv, exists := advertisementStore[id]
	return adv, exists
}

func (adv *Advertisement) NewPeer(u *User, race int32, team int32) *Peer {
	if isPeer(u) {
		// Ignore already added peers (via host & join)
		return nil
	}
	peer := &Peer{
		advertisement: adv,
		user:          u,
		race:          race,
		team:          team,
		invites:       make(map[*User]interface{}),
	}
	userId := peer.user.GetId()
	peerLock.Lock(userId)
	defer peerLock.Unlock(userId)
	adv.peers.Set(peer.user, peer)
	addPeer(u)
	return peer
}

func (adv *Advertisement) RemovePeer(user *User) {
	id := user.GetId()
	peerLock.Lock(id)
	adv.peers.Delete(user)
	removePeer(user)
	peerLock.Unlock(id)
	advLock.Lock(adv.id)
	if adv.peers.Len() == 0 || adv.host.GetUser() == user {
		advLock.Unlock(adv.id)
		adv.Delete()
	} else {
		advLock.Unlock(adv.id)
	}
}

func (adv *Advertisement) UpdatePeer(user *User, race int32, team int32) {
	userId := user.GetId()
	peerLock.Lock(userId)
	defer peerLock.Unlock(userId)
	peer, _ := adv.peers.Get(user)
	peer.race = race
	peer.team = team
}

func (adv *Advertisement) Delete() {
	advLock.Lock(adv.id)
	defer advLock.Unlock(adv.id)
	delete(advertisementStore, adv.id)
	host := adv.host
	hostId := host.GetUser().GetId()
	hostLock.Lock(hostId)
	removeHost(adv.host)
	hostLock.Unlock(hostId)
	for el := adv.peers.Oldest(); el != nil; el = el.Next() {
		u := el.Key
		id := u.GetId()
		peerLock.Lock(id)
		removePeer(u)
		peerLock.Unlock(id)
	}
}

func (adv *Advertisement) UpdateState(state int8) {
	advLock.Lock(adv.id)
	defer advLock.Unlock(adv.id)
	previousState := adv.state
	adv.state = state
	if adv.state == 1 && previousState != 1 {
		adv.startTime = time.Now().UTC().Unix()
		adv.visible = false
		adv.joinable = false
	}
}

func (adv *Advertisement) EncodePeers() i.A {
	var peers = make(i.A, adv.peers.Len())
	j := 0
	for el := adv.peers.Oldest(); el != nil; el = el.Next() {
		p := el.Value
		userId := el.Key.GetId()
		peerLock.RLock(userId)
		peers[j] = p.Encode()
		peerLock.RUnlock(userId)
		j++
	}
	return peers
}

func (adv *Advertisement) Encode() i.A {
	var visible uint8
	advLock.RLock(adv.id)
	defer advLock.RUnlock(adv.id)
	if adv.visible {
		visible = 1
	} else {
		visible = 0
	}
	var passworded uint8
	if adv.password.enabled {
		passworded = 1
	} else {
		passworded = 0
	}
	var startTime *int64
	if adv.startTime != 0 {
		startTime = &adv.startTime
	} else {
		startTime = nil
	}
	var started uint8
	if startTime != nil {
		started = 1
	} else {
		started = 0
	}
	return i.A{
		adv.id,
		adv.platformSessionId,
		0,
		"",
		"",
		"0",
		adv.host.GetUser().GetId(),
		started,
		adv.description,
		adv.description,
		visible,
		adv.mapName,
		adv.options,
		passworded,
		adv.maxPlayers,
		adv.slotInfo,
		adv.matchType,
		adv.EncodePeers(),
		0,
		0,
		0,
		0,
		1,
		1,
		startTime,
		adv.relayRegion,
		nil,
	}
}

func FindAdvertisements(matches func(adv *Advertisement) bool) []*Advertisement {
	var advs []*Advertisement
	for _, adv := range advertisementStore {
		advLock.RLock(adv.id)
		if matches(adv) {
			advs = append(advs, adv)
		}
		advLock.RUnlock(adv.id)
	}
	return advs
}
func FindAdvertisementsEncoded(matches func(adv *Advertisement) bool) []i.A {
	var advs []i.A
	advsOriginal := FindAdvertisements(matches)
	for _, adv := range advsOriginal {
		advLock.RLock(adv.id)
		advs = append(advs, adv.Encode())
		advLock.RUnlock(adv.id)
	}
	return advs
}

func IsInAdvertisement(user *User) bool {
	return IsHost(user) || IsPeer(user)
}

func IsPeer(user *User) bool {
	peerLock.RLock(user.GetId())
	defer peerLock.RUnlock(user.GetId())
	return isPeer(user)
}

func IsHost(user *User) bool {
	id := user.GetId()
	hostLock.RLock(id)
	defer hostLock.RUnlock(id)
	return isHost(user)
}

func isPeer(user *User) bool {
	_, exists := peers[user]
	return exists
}

func isHost(user *User) bool {
	_, exists := hosts[user]
	return exists
}

func addPeer(user *User) {
	if !isPeer(user) {
		peers[user] = struct{}{}
	}
}

func removePeer(user *User) {
	if isPeer(user) {
		delete(peers, user)
	}
}

func addHost(peer *Peer) {
	u := peer.GetUser()
	if !isHost(u) {
		hosts[u] = struct{}{}
	}
}

func removeHost(peer *Peer) {
	u := peer.GetUser()
	if isHost(u) {
		delete(hosts, u)
	}
}
