package httpauth

import (
    "bytes"
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"
    _ "github.com/mattn/go-sqlite3"
    "os"
    "testing"
)

var (
    sb         SqlAuthBackend
    mysqlDriverName = "mysql"
    mysqlDriverInfo = "travis@tcp(127.0.0.1:3306)/httpauth_test"
    pgDriverName = "postgres"
    pgDriverInfo = "user=postgres password='' dbname=httpauth_test sslmode=disable"
    sqliteDriverName = "sqlite3"
    sqliteDriverInfo = "./httpauth_test_sqlite.db"
)

//
// mysql tests
//

func TestSqlInit_mysql(t *testing.T) {
    con, err := sql.Open(mysqlDriverName, mysqlDriverInfo)
    if err != nil {
        t.Errorf("Couldn't set up test mysql database: %v", err)
        fmt.Printf("Couldn't set up test mysql database: %v\n", err)
        os.Exit(1)
    }
    err = con.Ping()
    if err != nil {
        t.Errorf("Couldn't ping test mysql database: %v", err)
        fmt.Printf("Couldn't ping test mysql database: %v\n", err)
        // t.Errorf("Couldn't ping test database: %v\n", err)
        os.Exit(1)
    }
    con.Exec("drop table goauth")
}

func TestNewSqlAuthBackend_mysql(t *testing.T) {
    var err error
    _, err = NewSqlAuthBackend(mysqlDriverName, mysqlDriverName + "_test")
    if err == nil {
        t.Fatal("Expected error on invalid connection.")
    }
    sb, err = NewSqlAuthBackend(mysqlDriverName, mysqlDriverInfo)
    if err != nil {
        t.Fatal(err.Error())
    }
    if sb.driverName != mysqlDriverName {
        t.Fatal("Driver name.")
    }
    if sb.dataSourceName != mysqlDriverInfo {
        t.Fatal("Driver info not saved.")
    }
}

func TestSqlAuthorizer_mysql(t *testing.T) {
    roles := make(map[string]Role)
    roles["user"] = 40
    roles["admin"] = 80
    _, err := NewAuthorizer(sb, []byte("testkey"), "user", roles)
    if err != nil {
        t.Fatal(err)
    }
}

func TestSaveUser_sql_mysql(t *testing.T) {
    user2 := UserData{"username2", "email2", []byte("passwordhash2"), "role2"}
    if err := sb.SaveUser(user2); err != nil {
        t.Fatalf("SaveUser sql error: %v", err)
    }

    user := UserData{"username", "email", []byte("passwordhash"), "role"}
    if err := sb.SaveUser(user); err != nil {
        t.Fatalf("SaveUser sql error: %v", err)
    }
}

func TestNewSqlAuthBackend_existing_mysql(t *testing.T) {
    b2, err := NewSqlAuthBackend(mysqlDriverName, mysqlDriverInfo)
    if err != nil {
        t.Fatal(err.Error())
    }

    user, err := b2.User("username")
    if err != nil {
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

func TestUser_existing_sql_mysql(t *testing.T) {
    if user, err := sb.User("username"); err == nil {
        if user.Username != "username" {
            t.Error("Username not correct.")
        }
        if user.Email != "email" {
            t.Error("User email not correct.")
        }
        if !bytes.Equal(user.Hash, []byte("passwordhash")) {
            t.Error("User password not correct.")
        }
    } else {
        t.Errorf("User not found: %v", err)
    }
    if user, err := sb.User("username2"); err == nil {
        if user.Username != "username2" {
            t.Error("Username not correct.")
        }
        if user.Email != "email2" {
            t.Error("User email not correct.")
        }
        if !bytes.Equal(user.Hash, []byte("passwordhash2")) {
            t.Error("User password not correct.")
        }
    } else {
        t.Fatalf("User not found: %v", err)
    }
}

func TestUser_notexisting_sql_mysql(t *testing.T) {
    if _, err := sb.User("notexist"); err != ErrMissingUser {
        t.Fatal("Not existing user found.")
    }
}

func TestUsers_sql_mysql(t *testing.T) {
    var (
        u1 UserData
        u2 UserData
    )
    users, err := sb.Users()
    if err != nil {
        t.Fatal(err.Error())
    }
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
        t.Error("Username not correct.")
    }
    if u1.Email != "email" {
        t.Error("User email not correct.")
    }
    if !bytes.Equal(u1.Hash, []byte("passwordhash")) {
        t.Error("User password not correct.")
    }
    if u2.Username != "username2" {
        t.Error("Username not correct.")
    }
    if u2.Email != "email2" {
        t.Error("User email not correct.")
    }
    if !bytes.Equal(u2.Hash, []byte("passwordhash2")) {
        t.Error("User password not correct.")
    }
}

func TestUpdateUser_sql_mysql(t *testing.T) {
    user2 := UserData{"username", "newemail", []byte("newpassword"), "newrole"}
    if err := sb.SaveUser(user2); err != nil {
        t.Fatalf("SaveUser sql error: %v", err)
    }
    u2, err := sb.User("username")
    if err != nil {
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

func TestSqlDeleteUser_sql_mysql(t *testing.T) {
    if err := sb.DeleteUser("username"); err != nil {
        t.Fatalf("DeleteUser error: %v", err)
    }
    err := sb.DeleteUser("username")
    if err == nil {
        t.Fatalf("DeleteUser should have raised error")
    } else if err != ErrDeleteNull {
        t.Fatalf("DeleteUser raised unexpected error: %v", err)
    }
}

func TestSqlReopen_mysql(t *testing.T) {
    var err error

    sb.Close()

    sb, err = NewSqlAuthBackend(mysqlDriverName, mysqlDriverInfo)
    if err != nil {
        t.Fatal(err.Error())
    }

    sb.Close()

    sb, err = NewSqlAuthBackend(mysqlDriverName, mysqlDriverInfo)
    if err != nil {
        t.Fatal(err.Error())
    }

    users, err := sb.Users()
    if err != nil {
        t.Fatal(err.Error())
    }
    if len(users) != 1 {
        t.Fatal("Users not loaded.")
    }
    if users[0].Username != "username2" {
        t.Error("Username not correct.")
    }
    if users[0].Email != "email2" {
        t.Error("User email not correct.")
    }
    if !bytes.Equal(users[0].Hash, []byte("passwordhash2")) {
        t.Error("User password not correct.")
    }
}

func TestSqlDelete2_mysql(t *testing.T) {
    if err := sb.DeleteUser("username2"); err != nil {
        t.Fatalf("DeleteUser error: %v", err)
    }
}

func TestSqlClose_mysql(t *testing.T) {
    sb.Close()
}

//
// postgres tests
//

func TestSqlInit_postgres(t *testing.T) {
    con, err := sql.Open(pgDriverName, pgDriverInfo)
    if err != nil {
        t.Errorf("Couldn't set up test postgres database: %v", err)
        fmt.Printf("Couldn't set up test postgres database: %v\n", err)
        os.Exit(1)
    }
    err = con.Ping()
    if err != nil {
        t.Errorf("Couldn't ping test postgres database: %v", err)
        fmt.Printf("Couldn't ping test postgres database: %v\n", err)
        // t.Errorf("Couldn't ping test database: %v\n", err)
        os.Exit(1)
    }
    con.Exec("drop table goauth")
}

func TestNewSqlAuthBackend_postgres(t *testing.T) {
    var err error
    _, err = NewSqlAuthBackend(pgDriverName, pgDriverName + "_test")
    if err == nil {
        t.Fatal("Expected error on invalid connection.")
    }
    sb, err = NewSqlAuthBackend(pgDriverName, pgDriverInfo)
    if err != nil {
        t.Fatal(err.Error())
    }
    if sb.driverName != pgDriverName {
        t.Fatal("Driver name.")
    }
    if sb.dataSourceName != pgDriverInfo{
        t.Fatal("Driver info not saved.")
    }
}

func TestSqlAuthorizer_postgres(t *testing.T) {
    roles := make(map[string]Role)
    roles["user"] = 40
    roles["admin"] = 80
    _, err := NewAuthorizer(sb, []byte("testkey"), "user", roles)
    if err != nil {
        t.Fatal(err)
    }
}

func TestSaveUser_sql_postgres(t *testing.T) {
    user2 := UserData{"username2", "email2", []byte("passwordhash2"), "role2"}
    if err := sb.SaveUser(user2); err != nil {
        t.Fatalf("SaveUser sql error: %v", err)
    }

    user := UserData{"username", "email", []byte("passwordhash"), "role"}
    if err := sb.SaveUser(user); err != nil {
        t.Fatalf("SaveUser sql error: %v", err)
    }
}

func TestNewSqlAuthBackend_existing_postgres(t *testing.T) {
    b2, err := NewSqlAuthBackend(pgDriverName, pgDriverInfo)
    if err != nil {
        t.Fatal(err.Error())
    }

    user, err := b2.User("username")
    if err != nil {
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

func TestUser_existing_sql_postgres(t *testing.T) {
    if user, err := sb.User("username"); err == nil {
        if user.Username != "username" {
            t.Error("Username not correct.")
        }
        if user.Email != "email" {
            t.Error("User email not correct.")
        }
        if !bytes.Equal(user.Hash, []byte("passwordhash")) {
            t.Error("User password not correct.")
        }
    } else {
        t.Errorf("User not found: %v", err)
    }
    if user, err := sb.User("username2"); err == nil {
        if user.Username != "username2" {
            t.Error("Username not correct.")
        }
        if user.Email != "email2" {
            t.Error("User email not correct.")
        }
        if !bytes.Equal(user.Hash, []byte("passwordhash2")) {
            t.Error("User password not correct.")
        }
    } else {
        t.Fatalf("User not found: %v", err)
    }
}

func TestUser_notexisting_sql_postgres(t *testing.T) {
    if _, err := sb.User("notexist"); err != ErrMissingUser {
        t.Fatal("Not existing user found.")
    }
}

func TestUsers_sql_postgres(t *testing.T) {
    var (
        u1 UserData
        u2 UserData
    )
    users, err := sb.Users()
    if err != nil {
        t.Fatal(err.Error())
    }
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
        t.Error("Username not correct.")
    }
    if u1.Email != "email" {
        t.Error("User email not correct.")
    }
    if !bytes.Equal(u1.Hash, []byte("passwordhash")) {
        t.Error("User password not correct.")
    }
    if u2.Username != "username2" {
        t.Error("Username not correct.")
    }
    if u2.Email != "email2" {
        t.Error("User email not correct.")
    }
    if !bytes.Equal(u2.Hash, []byte("passwordhash2")) {
        t.Error("User password not correct.")
    }
}

func TestUpdateUser_sql_postgres(t *testing.T) {
    user2 := UserData{"username", "newemail", []byte("newpassword"), "newrole"}
    if err := sb.SaveUser(user2); err != nil {
        t.Fatalf("SaveUser sql error: %v", err)
    }
    u2, err := sb.User("username")
    if err != nil {
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

func TestSqlDeleteUser_sql_postgres(t *testing.T) {
    if err := sb.DeleteUser("username"); err != nil {
        t.Fatalf("DeleteUser error: %v", err)
    }
    err := sb.DeleteUser("username")
    if err == nil {
        t.Fatalf("DeleteUser should have raised error")
    } else if err != ErrDeleteNull {
        t.Fatalf("DeleteUser raised unexpected error: %v", err)
    }
}

func TestSqlReopen_postgres(t *testing.T) {
    var err error

    sb.Close()

    sb, err = NewSqlAuthBackend(pgDriverName, pgDriverInfo)
    if err != nil {
        t.Fatal(err.Error())
    }

    sb.Close()

    sb, err = NewSqlAuthBackend(pgDriverName, pgDriverInfo)
    if err != nil {
        t.Fatal(err.Error())
    }

    users, err := sb.Users()
    if err != nil {
        t.Fatal(err.Error())
    }
    if len(users) != 1 {
        t.Fatal("Users not loaded.")
    }
    if users[0].Username != "username2" {
        t.Error("Username not correct.")
    }
    if users[0].Email != "email2" {
        t.Error("User email not correct.")
    }
    if !bytes.Equal(users[0].Hash, []byte("passwordhash2")) {
        t.Error("User password not correct.")
    }
}

func TestSqlDelete2_postgres(t *testing.T) {
    if err := sb.DeleteUser("username2"); err != nil {
        t.Fatalf("DeleteUser error: %v", err)
    }
}

func TestSqlClose_postgres(t *testing.T) {
    sb.Close()
}

//
// sqlite3 tests
//

func TestSqlInit_sqlite3(t *testing.T) {
    con, err := sql.Open(sqliteDriverName, sqliteDriverInfo)
    if err != nil {
        t.Errorf("Couldn't set up test sqlite3 database: %v", err)
        fmt.Printf("Couldn't set up test sqlite3 database: %v\n", err)
        os.Exit(1)
    }
    err = con.Ping()
    if err != nil {
        t.Errorf("Couldn't ping test sqlite3 database: %v", err)
        fmt.Printf("Couldn't ping test sqlite3 database: %v\n", err)
        // t.Errorf("Couldn't ping test database: %v\n", err)
        os.Exit(1)
    }
    con.Exec("drop table goauth")
}

func TestNewSqlAuthBackend_sqlite3(t *testing.T) {
    var err error
    sb, err = NewSqlAuthBackend(sqliteDriverName, sqliteDriverInfo)
    if err != nil {
        t.Fatal(err.Error())
        os.Exit(1)
    }
    if sb.driverName != sqliteDriverName {
        t.Fatal("Driver name.")
    }
    if sb.dataSourceName != sqliteDriverInfo {
        t.Fatal("Driver info not saved.")
    }
}

func TestSqlAuthorizer_sqlite3(t *testing.T) {
    roles := make(map[string]Role)
    roles["user"] = 40
    roles["admin"] = 80
    _, err := NewAuthorizer(sb, []byte("testkey"), "user", roles)
    if err != nil {
        t.Fatal(err)
    }
}

func TestSaveUser_sql_sqlite3(t *testing.T) {
    user2 := UserData{"username2", "email2", []byte("passwordhash2"), "role2"}
    if err := sb.SaveUser(user2); err != nil {
        t.Fatalf("SaveUser sql error: %v", err)
    }

    user := UserData{"username", "email", []byte("passwordhash"), "role"}
    if err := sb.SaveUser(user); err != nil {
        t.Fatalf("SaveUser sql error: %v", err)
    }
}

func TestNewSqlAuthBackend_existing_sqlite3(t *testing.T) {
    b2, err := NewSqlAuthBackend(sqliteDriverName, sqliteDriverInfo)
    if err != nil {
        t.Fatal(err.Error())
    }

    user, err := b2.User("username")
    if err != nil {
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

func TestUser_existing_sql_sqlite3(t *testing.T) {
    if user, err := sb.User("username"); err == nil {
        if user.Username != "username" {
            t.Error("Username not correct.")
        }
        if user.Email != "email" {
            t.Error("User email not correct.")
        }
        if !bytes.Equal(user.Hash, []byte("passwordhash")) {
            t.Error("User password not correct.")
        }
    } else {
        t.Errorf("User not found: %v", err)
    }
    if user, err := sb.User("username2"); err == nil {
        if user.Username != "username2" {
            t.Error("Username not correct.")
        }
        if user.Email != "email2" {
            t.Error("User email not correct.")
        }
        if !bytes.Equal(user.Hash, []byte("passwordhash2")) {
            t.Error("User password not correct.")
        }
    } else {
        t.Fatalf("User not found: %v", err)
    }
}

func TestUser_notexisting_sql_sqlite3(t *testing.T) {
    if _, err := sb.User("notexist"); err != ErrMissingUser {
        t.Fatal("Not existing user found.")
    }
}

func TestUsers_sql_sqlite3(t *testing.T) {
    var (
        u1 UserData
        u2 UserData
    )
    users, err := sb.Users()
    if err != nil {
        t.Fatal(err.Error())
    }
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
        t.Error("Username not correct.")
    }
    if u1.Email != "email" {
        t.Error("User email not correct.")
    }
    if !bytes.Equal(u1.Hash, []byte("passwordhash")) {
        t.Error("User password not correct.")
    }
    if u2.Username != "username2" {
        t.Error("Username not correct.")
    }
    if u2.Email != "email2" {
        t.Error("User email not correct.")
    }
    if !bytes.Equal(u2.Hash, []byte("passwordhash2")) {
        t.Error("User password not correct.")
    }
}

func TestUpdateUser_sql_sqlite3(t *testing.T) {
    user2 := UserData{"username", "newemail", []byte("newpassword"), "newrole"}
    if err := sb.SaveUser(user2); err != nil {
        t.Fatalf("SaveUser sql error: %v", err)
    }
    u2, err := sb.User("username")
    if err != nil {
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

func TestSqlDeleteUser_sql_sqlite3(t *testing.T) {
    if err := sb.DeleteUser("username"); err != nil {
        t.Fatalf("DeleteUser error: %v", err)
    }
    err := sb.DeleteUser("username")
    if err == nil {
        t.Fatalf("DeleteUser should have raised error")
    } else if err != ErrDeleteNull {
        t.Fatalf("DeleteUser raised unexpected error: %v", err)
    }
}

func TestSqlReopen_sqlite3(t *testing.T) {
    var err error

    sb.Close()

    sb, err = NewSqlAuthBackend(sqliteDriverName, sqliteDriverInfo)
    if err != nil {
        t.Fatal(err.Error())
    }

    sb.Close()

    sb, err = NewSqlAuthBackend(sqliteDriverName, sqliteDriverInfo)
    if err != nil {
        t.Fatal(err.Error())
    }

    users, err := sb.Users()
    if err != nil {
        t.Fatal(err.Error())
    }
    if len(users) != 1 {
        t.Fatal("Users not loaded.")
    }
    if users[0].Username != "username2" {
        t.Error("Username not correct.")
    }
    if users[0].Email != "email2" {
        t.Error("User email not correct.")
    }
    if !bytes.Equal(users[0].Hash, []byte("passwordhash2")) {
        t.Error("User password not correct.")
    }
}

func TestSqlDelete2_sqlite3(t *testing.T) {
    if err := sb.DeleteUser("username2"); err != nil {
        t.Fatalf("DeleteUser error: %v", err)
    }
}

func TestSqlClose_sqlite3(t *testing.T) {
    sb.Close()
    os.Remove(sqliteDriverInfo)
}
