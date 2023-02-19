package abstract_definition

// Session provides the only four operations of sessions, implementing them guarantees the implementation
// of a correct session.
type Session interface {
	SetValue(key, value interface{}) error
	GetValue(key interface{}) interface{}
	DeleteValue(key interface{}) error
	GetSessionId() string
}
