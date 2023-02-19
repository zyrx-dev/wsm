<h1>WSM (Web Sessions Manager)</h1>
<h2>Introduction:</h2>
Web Sessions Manager provides a Go sessions manager to add
to the web application in order to manage sessions with handling
them with ease and storing them in different storage media.

<h2>How to install?</h2>

```
go get github.com/zyrx-dev/wsm
```

<h2>Storage media available</h2>
Up until now, storage media supported by this package is <b>Memory</b>.

<h2>Upcoming storage media support</h2>
In the next version more storage media will be supported including:
1. JSON Files.
2. Postgres Database.

* <h2>Memory storage media</h2>
<h3>How to use it?</h3>

```
var sessionManager *wsm.SessionManager

func init() {
    sessionManager, _ = wsm.NewSessionManager("memory", cookieName as string, maxLiftime as int64)
    go sessionManager.SessionsExpirationRoutine()
}
```

<b>NOTE</b>: SessionExpirationRoutine() is a recursive method used to destroy sessions when they
are expired.


Then you are able to access the methods of session manager:

```
// to initialize a session
session, err := sessionManager.StartSession(response, request)

// to reset a session
sessionManager.EndSession(response, request)
```

After successfully initializing a session, you can retrieve/modify its value,
which it is a map of key/value pairs of type interface:

```
// to set a session value
err := session.SetValue("username", "zyrx")
// to retrieve a value
value := session.GetValue("username")
// to delete a value
err = session.DeleteValue("username")
// to retrieve current session id
id := session.GetSessionId()
```
