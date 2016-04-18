package main

import (
  // "log"
  "errors"
  "github.com/boltdb/bolt"
  // "github.com/agl/ed25519"
  "encoding/json"
)

type Database struct {
  Db *bolt.DB
  UsersCount int
}

func OpenDatabase(dbPath string) (*Database, error)  {
  db, err := bolt.Open(dbPath, 0600, nil)
  if err != nil {
    return nil, err
  }

  err = db.Update(func(tx *bolt.Tx) error {
    _, err := tx.CreateBucketIfNotExists([]byte("users"))

    return err
  })

  if err != nil {
    return nil, err
  }

  instance := new(Database)
  instance.Db = db
  instance.UsersCount = instance.CountUsers()

  return instance, nil
}

func (self *Database) Close() {
  self.Db.Close()
}

type User struct {
  // User name
  Name string `json:"name"`
  // User role admin, user, etc.
  Role string `json:"role"`
  // User keys
  Keys []UserKey `json:"keys"`
}

type UserKey struct {
  // Type could be ed25519 only
  Type string `json:"type",omitempty`
  // Key name home work etc
  Name string `json:"name"`
  // Key value
  Value string `json:"value"`
}

func (user *User) VerifySignature(message []byte, signature[]byte)  {

}

var DefaultUserRole string = "user"

var NotFound error = errors.New("not_found")
var DataDuplicate error = errors.New("data_duplicate")

func NewUser(name string, role string) (*User) {
  user := new(User)
  if len(role) < 1 {
    role = DefaultUserRole
  }

  user.Name = name
  user.Role = role

  return user
}

func (self *Database) GetUser(name string) (User, error) {
  var result []byte
  var user User

  err := self.Db.View(func (tx *bolt.Tx) error {
    bucket := tx.Bucket([]byte("users"))

    result = bucket.Get([]byte(name))

    return nil
  })

  if err != nil {
    return user, err
  }

  if result == nil {
    return user, NotFound
  }

  json.Unmarshal(result, &user)

  return user, nil
}

func (self *Database) AddUser(user User) (error) {
  err := self.Db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("users"))

    data, err := json.Marshal(user)
    if err != nil {
      return err
    }

    // TODO Emit error if user already exists
    return b.Put([]byte(user.Name), data)
  })

  if err != nil {
    return err
  }

  self.UsersCount += 1
  return nil
}

func (self *Database) RemoveUser(username string) error {
  var count int
  err := self.Db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("users"))

    key := []byte(username)

    if b.Get(key) != nil {
      b.Delete(key)
      count += 1
    }

    return nil
  })

  if err != nil {
    return err
  }

  self.UsersCount -= count
  return nil
}

func (self *Database) HasUser(name string) (bool) {
  var result bool

  self.Db.View(func (tx *bolt.Tx) error {
    bucket := tx.Bucket([]byte("users"))
    result = (bucket.Get([]byte(name)) == nil)

    return nil
  })

  return result
}

// Count users in the database
func (self * Database) CountUsers() (int) {
  var count int
  self.Db.View(func(tx *bolt.Tx) error {
    // Assume bucket exists and has keys
    b := tx.Bucket([]byte("users"))

    c := b.Cursor()

    for k, _ := c.First(); k != nil; k, _ = c.Next() {
        count += 1
    }

    return nil
  })

  return count
}
