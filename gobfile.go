package httpauth

import (
    "encoding/gob"
    "errors"
    "os"
    "fmt"
)

// GobFileAuthBackend stores user data and the location of the gob file.
type GobFileAuthBackend struct {
    filepath string
    users    map[string]UserData
}

// NewGobFileAuthBackend initializes a new backend by loading a map of users
// from a file.
// If the file doesn't exist, returns ErrMissingBackend.
func NewGobFileAuthBackend(filepath string) (b GobFileAuthBackend, e error) {
    b.filepath = filepath
    if _, err := os.Stat(b.filepath); err == nil {
        f, err := os.Open(b.filepath)
        defer f.Close()
        if err != nil {
            return b, fmt.Errorf("gobfilebackend: %v", err.Error())
        }
        dec := gob.NewDecoder(f)
        dec.Decode(&b.users)
    } else if !os.IsNotExist(err) {
        return b, fmt.Errorf("gobfilebackend: %v", err.Error())
    } else {
        return b, ErrMissingBackend
    }
    if b.users == nil {
        b.users = make(map[string]UserData)
    }
    return b, nil
}

// User returns the user with the given username.
func (b GobFileAuthBackend) User(username string) (user UserData, ok bool) {
    if user, ok := b.users[username]; ok {
        return user, ok
    }
    return user, false
}

// Users returns a slice of all users.
func (b GobFileAuthBackend) Users() (us []UserData) {
    for _, user := range b.users {
        us = append(us, user)
    }
    return
}

// SaveUser adds a new user, replacing one with the same username, and saves a
// gob file.
func (b GobFileAuthBackend) SaveUser(user UserData) error {
    b.users[user.Username] = user
    err := b.save()
    return err
}

func (b GobFileAuthBackend) save() error {
    f, err := os.Create(b.filepath)
    defer f.Close()
    if err != nil {
        return errors.New("auth file can't be edited. Is the data folder there?")
    }
    enc := gob.NewEncoder(f)
    err = enc.Encode(b.users)
    return err
}

// DeleteUser removes a user, raising ErrDeleteNull if that user was missing.
func (b GobFileAuthBackend) DeleteUser(username string) error {
    if _, ok := b.User(username); !ok {
        return ErrDeleteNull
    }
    delete(b.users, username)
    err := b.save()
    return err
}

// Close cleans up the backend. Currently a no-op for gobfiles.
func (b GobFileAuthBackend) Close() {
    ;
}
