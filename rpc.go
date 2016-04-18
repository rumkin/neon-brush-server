package main

import (
  "log"
  "net/http"
  // "github.com/boltdb/bolt"
  "errors"
)

type RpcService struct {
  Db *Database
}

var ErrForbidden error = errors.New("forbidden")
var ErrBadRequest error = errors.New("bad_request")
var ErrUnknown error = errors.New("unknown_error")
var ErrWrongRole error = errors.New("wrong_role")

func (self * RpcService) Add(r *http.Request, args *User, reply *bool) error {
  if self.Db.UsersCount > 0 {
    actorLogin := r.Header.Get("X-Auth-User")

    if len(actorLogin) < 1 {
      return ErrBadRequest
    }

    actor, err := self.Db.GetUser(actorLogin)

    if err != nil {
      if err == NotFound {
        return ErrForbidden
      } else {
        // Internal db error (should not be presented)
        log.Println(err)
        return ErrUnknown
      }
    }

    if actor.Role != "admin" {
      // Forbidden
      return ErrForbidden
    }
  } else if args.Role != "admin" {
    // The first user should be an admin
    return errors.New("wrong_role")
  }

  err := self.Db.AddUser(*args)

  if err != nil {
    *reply = false
    return err
  }

  *reply = true
  return nil
}

type GetArgs struct {
  Name string `json:"name"`
}

func (self * RpcService) Get(r *http.Request, args *GetArgs, reply *User) error {
  actorLogin := r.Header.Get("X-Auth-User")

  if len(actorLogin) < 1 {
    return ErrBadRequest
  }

  actor, err := self.Db.GetUser(actorLogin)

  if err != nil {
    if err == NotFound {
      return ErrForbidden
    } else {
      // Internal db error (should not be presented)
      log.Println(err)
      return ErrUnknown
    }
  }

  if actor.Role != "admin" {
    // Forbidden
    return ErrForbidden
  }

  user, err := self.Db.GetUser(args.Name)

  if err != nil {
    if err == NotFound {
      return err
    } else {
      log.Println(err)
      return ErrUnknown
    }
  }

  *reply = user

  return nil
}

type DeleteArgs struct {
  Name string `json:"name"`
}

func (self * RpcService) Delete(r *http.Request, args *DeleteArgs, reply *bool) error {
  actorLogin := r.Header.Get("X-Auth-User")

  if len(actorLogin) < 1 {
    return ErrBadRequest
  }

  actor, err := self.Db.GetUser(actorLogin)

  if err != nil {
    if err == NotFound {
      return ErrForbidden
    } else {
      // Internal db error (should not be presented)
      log.Println(err)
      return ErrUnknown
    }
  }

  if actor.Role != "admin" {
    // Forbidden
    return ErrForbidden
  }

  err = self.Db.RemoveUser(args.Name)
  *reply = true

  return nil
}
