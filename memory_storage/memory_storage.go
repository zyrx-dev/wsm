package memory_storage

import (
	"errors"
	"local/zyrx/backup/abstract_definition"
	"sync"
	"time"
)

// MemorySession is a struct holding the core data of a session, its unique ID,
// last time it has been accessed, and its value.
type MemorySession struct {
	id             string
	lastAccessTime time.Time
	value          map[interface{}]interface{}
}

// SetValue is a method for Session that takes key, value arguments both of type interface{}
// to set the session's value, and then save this change to the registered storage media
// as well as updating the session's last access time.
func (session *MemorySession) SetValue(key, value interface{}) error {
	// update this session in the registered storage media.
	session.value[key] = value
	return nil
}

// GetValue is a method for Session that takes a key argument of type interface{}
// to retrieve the session's value if it exists, otherwise it returns nil.
// It retrieves the value from the provided storage media and updates the session's last access time.
func (session *MemorySession) GetValue(key interface{}) interface{} {
	return session.value[key]
}

// DeleteValue is a method for Session that takes a key argument of type interface{}
// and delete the session's value stored in the storage media as well as updating the
// session's last access time.
// It returns nil for error on a successful deletion, otherwise it returns that error.
func (session *MemorySession) DeleteValue(key interface{}) error {
	delete(session.value, key)
	return nil
}

// GetSessionId is a method for Session that retrieves the current session ID
// calling this method.
func (session *MemorySession) GetSessionId() string {
	return session.id
}

// MemoryStorage represents a memory storage media type to store sessions in.
type MemoryStorage struct {
	sync.Mutex
	activeSessions int64
	sessions       map[string]*MemorySession
	//sessionsList []sessions
}

// InitializeSession is a method for MemoryStorage that takes a session ID argument of type string
// creates a new session, add it to memory, increasing the total active sessions count, and then return that session.
func (memory *MemoryStorage) InitializeSession(sessionId string) abstract_definition.Session {
	memory.Lock()
	defer memory.Unlock()
	var sessionValue map[interface{}]interface{}
	newSession := MemorySession{
		id:             sessionId,
		lastAccessTime: time.Now(),
		value:          sessionValue,
	}
	memory.activeSessions += 1
	memory.sessions[sessionId] = &newSession
	return &newSession
}

// RetrieveSession is a method for MemoryStorage that takes session ID of type string as an argument
// and returns the session stored in memory that belongs to the given ID, if it doesn't exist
// it returns a wsm.SessionNotExists error.
func (memory *MemoryStorage) RetrieveSession(sessionId string) (abstract_definition.Session, error) {
	memory.Lock()
	defer memory.Unlock()
	session, sessionExists := memory.sessions[sessionId]
	if !sessionExists {
		return nil, abstract_definition.SessionNotExist
	}
	return session, nil
}

// UpdateSessionLastAccess is a method for MemoryStorage that updates the session's
// last access time when it's used
func (memory *MemoryStorage) UpdateSessionLastAccess(sessionId string) error {
	memory.Lock()
	defer memory.Unlock()
	if _, err := memory.RetrieveSession(sessionId); errors.Is(err, abstract_definition.SessionNotExist) {
		return err
	}
	memory.sessions[sessionId].lastAccessTime = time.Now()
	return nil
}

// DestroySession is an implemented-overridden method for MemoryStorage that deletes a session
// from memory storage if found, otherwise it returns an error.
func (memory *MemoryStorage) DestroySession(sessionId string) error {
	memory.Lock()
	defer memory.Unlock()
	if _, err := memory.RetrieveSession(sessionId); errors.Is(err, abstract_definition.SessionNotExist) {
		return err
	}
	delete(memory.sessions, sessionId)
	memory.activeSessions -= 1
	return nil
}

// TerminateSessionOnExpiration is an overridden-implemented method for MemoryStorage that deletes
// sessions from memory that has exceeded a passed maximum lifetime parameter of type int64.
func (memory *MemoryStorage) TerminateSessionOnExpiration(maxLifetime int64) {
	memory.Lock()
	defer memory.Unlock()
	for sessionId, session := range memory.sessions {
		if session.lastAccessTime.Unix()+maxLifetime > time.Now().Unix() {
			delete(memory.sessions, sessionId)
			memory.activeSessions -= 1
		}
	}
}
