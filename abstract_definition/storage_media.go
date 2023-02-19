package abstract_definition

import (
	"errors"
)

// SessionNotExist is an error used when a session does not exist in the storage media.
var SessionNotExist = errors.New("wsm: session does not exist")

// StorageMedia provides a way to correctly handle a session in a provided storage media.
// Implementing these functions guarantees correct session handling in a specified storage media type.
type StorageMedia interface {
	InitializeSession(sessionId string) Session
	RetrieveSession(sessionId string) (Session, error)
	UpdateSessionLastAccess(sessionId string) error
	DestroySession(sessionId string) error
	TerminateSessionOnExpiration(maxLifetime int64)
}
