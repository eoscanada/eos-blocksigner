package main

import (
	"net/http"
	"fmt"
	"encoding/hex"
	"github.com/eoscanada/eos-go/ecc"
	"encoding/json"
	"log"
	"flag"
	"github.com/eoscanada/eos-go"
)

var keysFile = flag.String("keys-file", "", "P2P socket connection")
var port = flag.Int("port", 6666, "listening port")

func main() {

	flag.Parse()

	keyBag := eos.NewKeyBag()
	err := keyBag.ImportFromFile(*keysFile)
	if err != nil {
		log.Fatal("Importing keys: ", err)
	}

	http.HandleFunc("/v1/wallet/sign_digest", func(w http.ResponseWriter, r *http.Request) {

		fmt.Print("Signing digest... ")

		var inputs []string
		if err := json.NewDecoder(r.Body).Decode(&inputs); err != nil {
			fmt.Println("sign_digest: error:", err)
			http.Error(w, "couldn't decode input", 500)
			return
		}

		digest, err := hex.DecodeString(inputs[0])
		if err != nil {
			fmt.Println("digest decode : error:", err)
			http.Error(w, "couldn't decode digest", 500)
		}

		pubKey, err := ecc.NewPublicKey(inputs[1])
		if err != nil {
			fmt.Println("public key : error:", err)
			http.Error(w, "couldn't decode public key", 500)
		}

		signed, err := keyBag.SignDigest(digest, pubKey)
		if err != nil {
			fmt.Println("signing : error:", err)
			http.Error(w, fmt.Sprintf("error signing: %s", err), 500)
			return
		}

		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(signed)

		fmt.Println("done")

	})

	address := "127.0.0.1"
	listeningOn := fmt.Sprintf("%s:%d", address, *port)
	fmt.Printf("Listening for wallet operations on %s\n", listeningOn)
	if err := http.ListenAndServe(fmt.Sprintf("%s", listeningOn), nil); err != nil {
		log.Printf("Failed listening on port %s: %s\n", listeningOn, err)
	}
}
