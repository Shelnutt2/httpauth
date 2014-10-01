package httpauth

import (
    "database/sql"
)

// SqlAuthBackend database and database connection information.
type SqlAuthBackend struct {
    driverName     string
    dataSourceName string
    db             *sql.DB
}

// NewSqlAuthBackend initializes a new backend by testing the database
// connection and making sure the storage table exists. The table is called
// goauth.
//
// This uses the databases/sql package to open a connection. Its parameters
// should match the sql.Open function. See
// http://golang.org/pkg/database/sql/#Open for more information.
//
// Be sure to import "database/sql" and your driver of choice. If you're not
// using sql for your own purposes, you'll need to use the underscore to import
// for side effects; see http://golang.org/doc/effective_go.html#blank_import.
func NewSqlAuthBackend(driverName, dataSourceName string) (b SqlAuthBackend) {
    b.driverName = driverName
    b.dataSourceName = dataSourceName
    db, err := sql.Open(driverName, dataSourceName)
    if err != nil {
        panic(err)
    }
    b.db = db
    db.Exec(`create table if not exists goauth (Username varchar(255), Email varchar(255), Hash varchar(255), Role varchar(255), primary key (Username))`)
    return b
}

// User returns the user with the given username.
func (b SqlAuthBackend) User(username string) (user UserData, ok bool) {
    row := b.db.QueryRow(`select Email, Hash, Role from goauth where Username=?`, username)
    err := row.Scan(&user.Email, &user.Hash, &user.Role)
    if err != nil {
        return user, false
    }
    user.Username = username
    return user, true
}

// Users returns a slice of all users.
func (b SqlAuthBackend) Users() (us []UserData) {
    rows, err := b.db.Query("select Username, Email, Hash, Role from goauth")
    if err != nil {
        panic(err)
    }
    var (
        username, email, role string
        hash                  []byte
    )
    for rows.Next() {
        err = rows.Scan(&username, &email, &hash, &role)
        if err != nil {
            panic(err)
        }
        us = append(us, UserData{username, email, hash, role})
    }
    return
}

// SaveUser adds a new user, replacing one with the same username.
func (b SqlAuthBackend) SaveUser(user UserData) (err error) {
    if _, ok := b.User(user.Username); !ok {
        _, err = b.db.Exec("insert into goauth (Username, Email, Hash, Role) values (?, ?, ?, ?)", user.Username, user.Email, user.Hash, user.Role)
    } else {
        _, err = b.db.Exec("update goauth set Email=?, Hash=?, Role=? where Username=?", user.Email, user.Hash, user.Role, user.Username)
    }
    return
}

// DeleteUser removes a user, raising ErrDeleteNull if that user was missing.
func (b SqlAuthBackend) DeleteUser(username string) error {
    result, err := b.db.Exec("delete from goauth where Username=?", username)
    if err != nil {
        return err
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return ErrDeleteNull
    }
    return err
}

// Close cleans up the backend by terminating the database connection.
func (b SqlAuthBackend) Close() {
    err := b.db.Close()
    if err != nil {
        panic(err)
    }
}
