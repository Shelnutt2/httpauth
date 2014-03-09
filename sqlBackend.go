package goauth

import (
    "database/sql"
    //"os"
    //"errors"
)

// SqlAuthBackend stores user data and the location of the gob file.
type SqlAuthBackend struct {
    driverName string
    dataSourceName string
    db *sql.DB
}

func (b SqlAuthBackend) connect() *sql.DB {
    con, err := sql.Open(b.driverName, b.dataSourceName)
    if err != nil {
        panic(err)
    }
    return con
}

// NewSqlAuthBackend initializes a new backend by loading a map of users
// from a file.
func NewSqlAuthBackend(driverName, dataSourceName string) (b SqlAuthBackend) {
    b.driverName = driverName
    b.dataSourceName = dataSourceName
    con := b.connect()
    con.Exec(`create table if not exists goauth (Username varchar(255), Email varchar(255), Hash varchar(255), primary key (Username))`)
    return b
}

// User returns the user with the given username.
func (b SqlAuthBackend) User(username string) (user UserData, ok bool) {
    con := b.connect()
    defer con.Close()
    row := con.QueryRow("select Email, Hash from goauth where Username=?", username)
    var (
        email string
        hash []byte
    )
    err := row.Scan(&email, &hash)
    if err != nil {
        return user, false
    }
    user.Username = username
    user.Email = email
    user.Hash = hash
    return user, true
}

// Users returns a slice of all users.
func (b SqlAuthBackend) Users() (us []UserData) {
    con := b.connect()
    rows, err := con.Query("select Username, Email, Hash from goauth")
    if err != nil { panic(err) }
    var (
        username, email string
        hash []byte
    )
    for rows.Next() {
        err = rows.Scan(&username, &email, &hash)
        if err != nil { panic(err) }
        us = append(us, UserData{username, email, hash})
    }
    return
}

// SaveUser adds a new user, replacing one with the same username, and saves a
// gob file.
func (b SqlAuthBackend) SaveUser(user UserData) error {
    con := b.connect()
    _, err := con.Exec("insert into goauth (Username, Email, Hash) values (?, ?, ?)", user.Username, user.Email, user.Hash)
    return err
}

// DeleteUser removes a user.
func (b SqlAuthBackend) DeleteUser(username string) error {
    con := b.connect()
    _, err := con.Exec("delete from goauth where Username=?", username)
    return err
}
