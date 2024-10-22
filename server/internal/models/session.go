package models

import (
	"github.com/luskaner/aoe2DELanServer/server/internal"
	"sync"
	"time"
)

type Session struct {
	id              string
	expiryTimer     *time.Timer
	userId          int32
	expiryTimerLock sync.Mutex
	gameId          string
}

var userIdSession = internal.NewSafeMap[int32, *Session]()
var sessionStore = internal.NewSafeMap[string, *Session]()

var (
	sessionLetters  = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	sessionDuration = 5 * time.Minute
)

func generateSessionId() string {
	sessionId := make([]rune, 30)
	for {
		for i := range sessionId {
			internal.RngLock.Lock()
			sessionId[i] = sessionLetters[internal.Rng.Intn(len(sessionLetters))]
			internal.RngLock.Unlock()
		}
		sessionIdStr := string(sessionId)
		if _, exists := GetSessionById(sessionIdStr); !exists {
			return sessionIdStr
		}
	}
}

func (sess *Session) GetId() string {
	return sess.id
}

func (sess *Session) GetUserId() int32 {
	return sess.userId
}

func (sess *Session) GetGameId() string {
	return sess.gameId
}

func CreateSession(gameId string, userId int32) string {
	session := &Session{
		id:     generateSessionId(),
		userId: userId,
		gameId: gameId,
	}
	session.expiryTimer = time.AfterFunc(sessionDuration, func() {
		session.Delete()
	})
	sessionStore.Store(session.id, session)
	userIdSession.Store(userId, session)
	return session.id
}

func (sess *Session) Delete() {
	_ = sess.expiryTimer.Stop()
	userIdSession.Delete(sess.userId)
	sessionStore.Delete(sess.id)
}

func (sess *Session) ResetExpiryTimer() {
	sess.expiryTimerLock.Lock()
	defer sess.expiryTimerLock.Unlock()
	if !sess.expiryTimer.Stop() {
		<-sess.expiryTimer.C
	}
	sess.expiryTimer.Reset(sessionDuration)
}

func GetSessionById(sessionId string) (*Session, bool) {
	return sessionStore.Load(sessionId)
}

func GetSessionByUserId(userId int32) (*Session, bool) {
	return userIdSession.Load(userId)
}
