package httpauth

import (
    "bytes"
    "fmt"
    "os"
    "testing"
    "gopkg.in/mgo.v2"
    //"gopkg.in/mgo.v2/bson"
)

var (
    backend         MongodbAuthBackend
    url = "mongodb://localhost/"
    db = "test"
)

func TestMongodbInit(t *testing.T) {
    con, err := mgo.Dial(url)
    if err != nil {
        t.Errorf("Couldn't set up test mongodb session: %v", err)
        fmt.Printf("Couldn't set up test mongodb session: %v\n", err)
        os.Exit(1)
    }
    err = con.Ping()
    if err != nil {
        t.Errorf("Couldn't ping test mongodb database: %v", err)
        fmt.Printf("Couldn't ping test mongodb database: %v\n", err)
        // t.Errorf("Couldn't ping test database: %v\n", err)
        os.Exit(1)
    }
    db := con.DB(db)
    err = db.DropDatabase()
    if err != nil {
        t.Errorf("Couldn't drop test mongodb database: %v", err)
        fmt.Printf("Couldn't drop test mongodb database: %v\n", err)
        // t.Errorf("Couldn't ping test database: %v\n", err)
        os.Exit(1)
    }
}

func TestNewMongodbAuthBackend(t *testing.T) {
    backend, err := NewMongodbBackend(url, db)
    if err != nil {
        t.Fatalf("NewMongodbBackend error: %v", err)
    }
    if backend.mongoUrl != url {
        t.Fatal("Url name.")
    }
    if backend.database != db {
        t.Fatal("DB not saved.")
    }
}

func TestSaveUser_mongodb(t *testing.T) {
    user2 := UserData{"username2", "email2", []byte("passwordhash2"), "role2"}
    if err := backend.SaveUser(user2); err != nil {
        t.Fatalf("SaveUser mongodb error: %v", err)
    }

    user := UserData{"username", "email", []byte("passwordhash"), "role"}
    if err := backend.SaveUser(user); err != nil {
        t.Fatalf("SaveUser mongodb error: %v", err)
    }
}

func TestNewMongodbAuthBackend_existing(t *testing.T) {
    b2, err := NewMongodbBackend(driverName, driverInfo)
    if err != nil {
        t.Fatalf("NewMongodbBackend (existing) error: %v", err)
    }

    user, ok := b2.User("username")
    if !ok {
        t.Fatal("Secondary backend failed")
    }
    if user.Username != "username" {
        t.Fatal("Username not correct.")
    }
    if user.Email != "email" {
        t.Fatal("User email not correct.")
    }
    if !bytes.Equal(user.Hash, []byte("passwordhash")) {
        t.Fatal("User password not correct.")
    }
}

func TestUser_existing_mongodb(t *testing.T) {
    if user, ok := backend.User("username"); ok {
        if user.Username != "username" {
            t.Fatal("Username not correct.")
        }
        if user.Email != "email" {
            t.Fatal("User email not correct.")
        }
        if !bytes.Equal(user.Hash, []byte("passwordhash")) {
            t.Fatal("User password not correct.")
        }
    } else {
        t.Fatal("User not found")
    }
    if user, ok := backend.User("username2"); ok {
        if user.Username != "username2" {
            t.Fatal("Username not correct.")
        }
        if user.Email != "email2" {
            t.Fatal("User email not correct.")
        }
        if !bytes.Equal(user.Hash, []byte("passwordhash2")) {
            t.Fatal("User password not correct.")
        }
    } else {
        t.Fatal("User not found")
    }
}

func TestUser_notexisting_mongodb(t *testing.T) {
    if _, ok := backend.User("notexist"); ok {
        t.Fatal("Not existing user found.")
    }
}

func TestUsers_mongodb(t *testing.T) {
    var (
        u1 UserData
        u2 UserData
    )
    users := backend.Users()
    if len(users) != 2 {
        t.Fatal("Wrong amount of users found.")
    }
    if users[0].Username == "username" {
        u1 = users[0]
        u2 = users[1]
    } else if users[1].Username == "username" {
        u1 = users[1]
        u2 = users[0]
    } else {
        t.Fatal("One of the users not found.")
    }

    if u1.Username != "username" {
        t.Fatal("Username not correct.")
    }
    if u1.Email != "email" {
        t.Fatal("User email not correct.")
    }
    if !bytes.Equal(u1.Hash, []byte("passwordhash")) {
        t.Fatal("User password not correct.")
    }
    if u2.Username != "username2" {
        t.Fatal("Username not correct.")
    }
    if u2.Email != "email2" {
        t.Fatal("User email not correct.")
    }
    if !bytes.Equal(u2.Hash, []byte("passwordhash2")) {
        t.Fatal("User password not correct.")
    }
}

func TestUpdateUser_mongodb(t *testing.T) {
    user2 := UserData{"username", "newemail", []byte("newpassword"), "newrole"}
    if err := backend.SaveUser(user2); err != nil {
        t.Fatalf("SaveUser mongodb error: %v", err)
    }
    u2, ok := backend.User("username")
    if !ok {
        t.Fatal("Updated user not found")
    }
    if u2.Username != "username" {
        t.Fatal("Username not correct.")
    }
    if u2.Email != "newemail" {
        t.Fatal("User email not correct.")
    }
    if u2.Role != "newrole" {
        t.Fatalf("User role not correct: found %v, expected %v", u2.Role, "newrole");
    }
    if !bytes.Equal(u2.Hash, []byte("newpassword")) {
        t.Fatal("User password not correct.")
    }
}

func TestMongodbDeleteUser_mongodb(t *testing.T) {
    if err := backend.DeleteUser("username"); err != nil {
        t.Fatalf("DeleteUser error: %v", err)
    }
    if err := backend.DeleteUser("username"); err != nil {
        t.Fatalf("DeleteUser error: %v", err)
    }

    if err := backend.DeleteUser("username2"); err != nil {
        t.Fatalf("DeleteUser error: %v", err)
    }
}

