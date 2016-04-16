package main

import (
  // "log"
  // "errors"
  "github.com/boltdb/bolt"
  "encoding/json"
)

type Database struct {
  Db *bolt.DB
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

var DefaultUserRole string = "user"

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

  if result == nil || err != nil {
    return user, err
  }

  json.Unmarshal(result, &user)

  return user, nil
}

func (self *Database) AddUser(user User) (error) {
  return self.Db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte("users"))

    data, err := json.Marshal(user)
    if err != nil {
      return err
    }

    // TODO Emit error if user already exists
    return b.Put([]byte(user.Name), data)
  })
}
