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
  "github.com/agl/ed25519"
  // "github.com/agl/ed25519/edwards25519"
)


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

type VerifyArgs struct {
  Username string `json:"username"`
  Signature string `json:"signature"`
  Type string `json:"type"`
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
      if err == NotFound {
          fmt.Fprintf(w, "User not found\n")
          return
      }

      fmt.Fprintf(os.Stderr, "Read error:", err.Error())
      fmt.Fprintf(w, "Fail")
      return
    }

    var key UserKey
    var keyFound bool
    for _,v := range user.Keys {
        if (v.Name == args.Type) {
            key = v
            keyFound = true
            break
        }
    }

    if keyFound != true {
        fmt.Fprint(w, "{\"status\": false}")
        return
    }

    var publicKey [32]byte
    var signature [64]byte

    copy(signature[0:64], args.Signature)
    copy(publicKey[0:32], key.Value)

    verifyData := make(map[string]interface{})
    verifyData["username"] = args.Username
    verifyData["type"] = args.Type
    verifyData["resource"] = req.Header.Get("Referrer")

    hash, err := GenerateHash(verifyData)
    if err != nil {
        fmt.Fprint(w, "{\"error\": \"unknown_error\"}")
        return
    }

    // keyBytes := make([]byte, 32)
    // copy(hash[0:32], hashBytes)
    if (! ed25519.Verify(&publicKey, hash, &signature)) {
        fmt.Fprint(w, "{\"status\": false}")
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

  startServer(connType, connAddr, dbPath, superUser)
}
