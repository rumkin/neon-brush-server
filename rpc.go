package main

import (
  // "log"
  "net/http"
  // "github.com/boltdb/bolt"
  // "errors"
)

type RpcService struct {
  Db *Database
}

func (self * RpcService) Add(r *http.Request, args *User, reply *bool) (error) {
  err := self.Db.AddUser(*args)

  if err != nil {
    *reply = false
    return err
  }

  *reply = true
  return nil
}
