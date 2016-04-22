package main

import (
    "encoding/json"
    "crypto/sha256"
)


func GenerateHash(value interface{}) ([]byte, error) {
    jsonString, err := json.Marshal(value)
    if err != nil {
        return nil, err
    }

    hash := sha256.New()
    hash.Write(jsonString)
    return hash.Sum(nil), nil
}
