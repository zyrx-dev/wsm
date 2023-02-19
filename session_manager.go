// Package wsm (Web Sessions Manager) provides a way to handle and manage web sessions,
// alongside multiple storage media such as memory, files, and postgres database, to store those sessions.
package wsm_backup

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"local/zyrx/backup/abstract_definition"
	"local/zyrx/backup/file_storage"
	"local/zyrx/backup/memory_storage"
	"local/zyrx/backup/postgres_storage"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SessionManager provides a general way to manage sessions by maintaining a unique session ID,
// keeping a single session per user, storing sessions in a supported storage media,
// handle sessions expiration through lifetimes and correct cleanup.
type SessionManager struct {
	sync.Mutex
	cookieName   string
	storageMedia abstract_definition.StorageMedia
	maxLifetime  int64
}

// supportedStorageMedia is a map of built-in storage media types mapped to a string key (indicator).
var supportedStorageMedia = map[string]abstract_definition.StorageMedia{
	"memory":   &memory_storage.MemoryStorage{},
	"file":     &file_storage.FileStorage{},
	"postgres": &postgres_storage.PostgresStorage{},
}

// supportedStorageMediaTypes is a slice of the currently supported storage media types in the package.
// Used to be displayed on the error of unsupported storage media type.
var supportedStorageMediaTypes = []string{"memory", "file", "postgres"}

// RegisteredStorageMedia is the storage media type that has already been used
type RegisteredStorageMedia struct {
	StorageMediaType string                           `json:"type"`
	StorageMedia     abstract_definition.StorageMedia `json:"storage-media"`
	//SessionType      Session      `json:"session-type"`
}

// sessionStorage checks for a json file holding the last registered storage media to retrieve it.
// If it exists and the storage media type passed as an argument match, it gets retrieved,
// otherwise a message prompts asking to confirm the replacement of the old storage with all its data
// with the new one.
// If the json file doesn't exist, it registers the provided storage media type if it is supported,
// if it's not supported it returns an error.
func sessionStorage(storageMediaType string, storageMedia abstract_definition.StorageMedia) (abstract_definition.StorageMedia, error) {
	fileMatches, err := filepath.Glob("registered_storage/*.json")
	if err != nil {
		log.Fatal(err)
	}
	if fileMatches != nil {
		fileName := strings.Split(strings.Split(fileMatches[0], "\\")[1], ".")[0]
		file, err := os.Open(fileMatches[0])
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		fileData, err := io.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}
		var registeredStorageMedia RegisteredStorageMedia
		err = json.Unmarshal(fileData, &registeredStorageMedia)
		if err != nil {
			log.Fatal(err)
		}
		if fileName != storageMediaType {
			fmt.Printf("Would you like to change storage type from %s to %s? ", fileName, storageMediaType)
			var answer string
			_, err = fmt.Scan(&answer)
			if err != nil {
				log.Fatal(err)
			}
			answer = strings.ToLower(answer)
			if answer == "yes" || answer == "y" {
				// ChangeStorageMedia(oldStorageType, newStorageType)
				// Here we use this function to perform the transformation of sessions data from the old
				// to the new storage media type.
			} else {
				return registeredStorageMedia.StorageMedia, nil
			}
		} else {
			return registeredStorageMedia.StorageMedia, nil
		}
	}
	registeredStorageMedia := RegisteredStorageMedia{
		StorageMediaType: storageMediaType,
		StorageMedia:     storageMedia,
	}
	jsonRepresentation, err := json.MarshalIndent(registeredStorageMedia, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Create(fmt.Sprintf("registered_storage/%s.json", storageMediaType))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	_, err = file.Write(jsonRepresentation)
	if err != nil {
		log.Fatal(err)
	}
	return storageMedia, nil
}

// NewSessionManager is a function that initializes a new SessionManager,
// setting its storage media to either memory, file, or postgres,
// the cookie it's going to be sent in, and its maximum lifetime.
// It returns an error in case the storage media type is not supported.
func NewSessionManager(storageMediaType, cookieName string, maxLifetime int64) (*SessionManager, error) {
	storageMediaType = strings.ToLower(storageMediaType)
	storageMedia, storageMediaSupported := supportedStorageMedia[storageMediaType]
	if !storageMediaSupported {
		errorMessage := fmt.Errorf("wsm: unsupported storage media type %v, "+
			"the supported storage media types are %v", storageMediaType, supportedStorageMediaTypes)
		return nil, errorMessage
	}
	registeredStorage, err := sessionStorage(storageMediaType, storageMedia)
	if err != nil {
		return nil, err
	}
	newSessionManager := &SessionManager{
		cookieName:   cookieName,
		storageMedia: registeredStorage,
		maxLifetime:  maxLifetime,
	}
	return newSessionManager, nil
}

// generateUniqueSessionID is a method for SessionManager used to generate a secure random number
// to serve as a unique session ID for newly created sessions.
func (manager *SessionManager) generateUniqueSessionID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// StartSession is a method for SessionManager used to initialize a session with a unique ID for a new user,
// generate and set the cookie with proper values.
// If the user already has a session, it gets retrieved based on their cookie info.
// Returns an error if session ID could not be read from cookie or the session could not be retrieved.
func (manager *SessionManager) StartSession(response http.ResponseWriter, request *http.Request) (abstract_definition.Session, error) {
	manager.Lock()
	defer manager.Unlock()
	var session abstract_definition.Session
	cookie, err := request.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		sessionId := manager.generateUniqueSessionID()
		session = manager.storageMedia.InitializeSession(sessionId)
		cookie = &http.Cookie{Name: manager.cookieName, Value: url.QueryEscape(sessionId), Path: "/",
			HttpOnly: true, MaxAge: int(manager.maxLifetime)}
		http.SetCookie(response, cookie)
	} else {
		sessionId, err := url.QueryUnescape(cookie.Value)
		if err != nil {
			return nil, err
		}
		session, err = manager.storageMedia.RetrieveSession(sessionId)
		if err != nil {
			return nil, err
		}
	}
	return session, nil
}

// EndSession is a method for SessionManager used to reset the user's session on their logout.
// It sets the cookie provided by previously set name in SessionManager, to expired values
// rendering the session in-active.
func (manager *SessionManager) EndSession(response http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}
	manager.Lock()
	defer manager.Unlock()
	err = manager.storageMedia.DestroySession(cookie.Value)
	expiration := time.Now()
	cookie = &http.Cookie{Name: manager.cookieName, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
	http.SetCookie(response, cookie)
}

// SessionsExpirationRoutine is a method for SessionManager, used as a go routine to terminate
// sessions after they pass their expiration date.
// It's called periodically after the set maximum lifetime value elapsed.
func (manager *SessionManager) SessionsExpirationRoutine() {
	manager.Lock()
	defer manager.Unlock()
	manager.storageMedia.TerminateSessionOnExpiration(manager.maxLifetime)
	time.AfterFunc(time.Duration(manager.maxLifetime), manager.SessionsExpirationRoutine)
}
