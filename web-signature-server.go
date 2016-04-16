package main

import (
  "log"
  "os"
  "fmt"
  "regexp"
  "net"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "github.com/gorilla/rpc"
  gJson "github.com/gorilla/rpc/json"
  // "errors"
  // "github.com/boltdb/bolt"
  // "github.com/agl/ed25519"
)

// var db *bolt.DB
//
// type Users struct {
// }
//
// type PutArgs struct {
//   Name string
//   PublicKey string
// }
//
// type GetArgs struct {
//   Name string
// }
//
// func (api *Users) Put(args *PutArgs, result *bool) error {
//   log.Println("Put");
//   if len(args.Name) < 1 {
//     return errors.New("Name is empty")
//   }
//
//   if len(args.PublicKey) != 32  {
//     return errors.New("PublicKey length should be 32")
//   }
//
//   db.Update(func (tx *bolt.Tx) error {
//     bucket := tx.Bucket([]byte("users"))
//
//     return bucket.Put([]byte(args.Name), []byte(args.PublicKey))
//   })
//
//   *result = true
//
//   return nil
// }
//
// func (api *Users) Get(args *GetArgs, result *string) error {
//   log.Println("Get");
//   if len(args.Name) < 1 {
//     return errors.New("Name is empty")
//   }
//
//   db.View(func (tx *bolt.Tx) error {
//     bucket := tx.Bucket([]byte("users"))
//
//     *result = string(bucket.Get([]byte(args.Name)))
//
//     return nil
//   })
//
//   return nil
// }

// func startServer (ctype string, caddr string) {
//   api := new(Users)
//
//   server := rpc.NewServer()
//   server.Register(api)
//   // server.HandleHTTP(jsonrpc.)
//
//   listener, err := net.Listen(ctype, caddr)
//
//   if err != nil {
//       log.Fatal("listen error:", err)
//   }
//
//   defer listener.Close()
//
//   for {
//       conn, err := listener.Accept()
//
//       if err != nil {
//           log.Fatal(err)
//       }
//
//       go server.ServeCodec(jsonrpc.NewServerCodec(conn))
//   }
// }
//
// func startClient(ctype string, caddr string) {
//   conn, err := net.Dial(ctype, caddr)
//
//   if err != nil {
//       panic(err)
//   }
//   defer conn.Close()
//
//   c := jsonrpc.NewClient(conn)
//
//   var reply bool
//   var args = PutArgs{"user", "12345678901234567890123456789012"}
//   err = c.Call("Users.Put", args, &reply)
//   if err != nil {
//       log.Fatal("arith error:", err)
//   }
//   log.Println("msg", reply)
// }
type VerifyArgs struct {
  Username string `json:"username"`
  Signature string `json:"signature"`
  Type string `json:"type"`
}

// startServer Starts http server instance with verification handler
func startServer(socket string, port string, dbPath string, superUser []string) (error) {
  log.Println("conn type", socket)
  log.Println("conn port", port)
  log.Println("dbPath", dbPath)

  for _,v := range superUser {
    log.Println("superUser: ", v)
  }

  listener, err := net.Listen(socket, port)
  if err != nil {
    log.Fatal("Listen error: ", err)
  }

  db, err := OpenDatabase(dbPath)
  if err != nil {
    log.Fatal("Db error:", err)
  }

  defer db.Close()

  // Handle http verification requests
  http.HandleFunc("/", HttpDatabaseHandler(db))

  rpcService := new(RpcService)
  rpcService.Db = db

  // Handle RPC users manage requests
  s := rpc.NewServer()
  s.RegisterCodec(gJson.NewCodec(), "application/json")
  s.RegisterService(rpcService, "Users")
  http.Handle("/rpc", s)

  http.Serve(listener, nil)

  return nil
}

// Handle
func HttpDatabaseHandler(db *Database) (func (http.ResponseWriter, *http.Request)) {
  return func (w http.ResponseWriter, req *http.Request){
    body, err := ioutil.ReadAll(req.Body)

    if err != nil {
      fmt.Fprintf(os.Stderr, "Read error:", err)
      fmt.Fprint(w, "Fail")
      return
    }


    args := VerifyArgs{}
    json.Unmarshal(body, &args)

    user, err := db.GetUser(args.Username)

    if err != nil {
      fmt.Fprintf(os.Stderr, "Read error:", err)
      fmt.Fprintf(w, "Fail")
      return
    }

    data, err := json.Marshal(user)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Json marshal:", err)
      fmt.Fprint(w, "FAIL")
      return
    }

    fmt.Fprint(w, string(data), "\n")
  }
}


func main() {
  if len(os.Args) < 2 {
    log.Fatal("port not specified")
  }

  var dbPath string
  if len(os.Args) < 3 {
    dbPath = "webrsa.db"
  } else {
    dbPath = os.Args[2]
  }

  var superUser []string
  if len(os.Args) < 4 {
    superUser = []string{"root"}
  } else {
    re := regexp.MustCompile("\\s*,\\s*")
    superUser = re.Split(os.Args[3], -1)
  }

  port := os.Args[1]
  var connType string
  var connAddr string
  isSocket, _ := regexp.MatchString("^\\d+$", port)
  if isSocket {
    connType = "tcp"
    connAddr = ":" + port
  } else {
    connType = "unix"
    connAddr = port
  }

  // var err error
  // db, err = bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
  //
  // if err != nil {
  //   panic(err)
  // }
  //
  // defer db.Close()
  //
  //
  // db.Update(func(tx *bolt.Tx) error {
  //   _, err := tx.CreateBucketIfNotExists([]byte("users"))
  //
  //   return err
  // })
  //
  // go startServer(connType, connAddr)
  //
  // time.Sleep(1 * time.Second)
  //
  // startClient(connType, connAddr)
  // time.Sleep(10 * time.Second)
  startServer(connType, connAddr, dbPath, superUser)
}
