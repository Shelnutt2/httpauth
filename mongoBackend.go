package httpauth

import (
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "fmt"
)


// MongodbAuthBackend stores database connection information.
type MongodbAuthBackend struct {
    mongoUrl string
    database string
    session  *mgo.Session
}

func (b MongodbAuthBackend) connect() *mgo.Collection {
    session := b.session.Copy()
    return session.DB(b.database).C("goauth")
}

// NewMongodbAuthBackend initializes a new backend.
func NewMongodbBackend(mongoUrl string, database string) (b MongodbAuthBackend) {
    b.mongoUrl = mongoUrl
    b.database = database
    session, err := mgo.Dial(b.mongoUrl)
    if err != nil {
        panic(err)
    }
    b.session = session
    return
}

// User returns the user with the given username.
func (b MongodbAuthBackend) User(username string) (user UserData, ok bool) {
    var result UserData

    c := b.connect()
    defer c.Database.Session.Close()

    err := c.Find(bson.M{"Username": username}).One(&result)
    if err != nil {
        return result, false
    }
    return result, true
}

// Users returns a slice of all users.
func (b MongodbAuthBackend) Users() (us []UserData) {
    var results []UserData

    c := b.connect()
    defer c.Database.Session.Close()

    err := c.Find(bson.M{}).All(&results)
    if err != nil {
        // TODO: Remove
        fmt.Printf("got an error finding a doc %v\n")
    }
    return results
}

// SaveUser adds a new user, replacing if the same username is in use.
func (b MongodbAuthBackend) SaveUser(user UserData) error {
    c := b.connect()
    defer c.Database.Session.Close()

    hash := string(user.Hash)
    m := c.Find(bson.M{ "Username": user.Username })
    l, err := m.Count()
    if err != nil {
        panic(err)
    }
    if (l == 0) {
        err = c.Insert(bson.M{ "Username": user.Username, "Hash": hash, "Email": user.Email, "Role": user.Role })
    } else {
        err = c.Update(bson.M{ "Username": user.Username }, bson.M{ "Username": user.Username, "Hash": hash, "Email": user.Email, "Role": user.Role })
    }
    return err
}

// DeleteUser removes a user. An error is raised if the user isn't found.
// TODO: Should that error be raised? (Different than sql)
func (b MongodbAuthBackend) DeleteUser(username string) error {
    c := b.connect()
    defer c.Database.Session.Close()

    err := c.Remove(bson.M{"Username": username})
    return err
}

func (b MongodbAuthBackend) Close() {
    b.session.Close()
}
