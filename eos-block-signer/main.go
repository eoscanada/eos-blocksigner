package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"os"

	"github.com/eoscanada/eos-go/ecc"
	eosvault "github.com/eoscanada/eosc/vault"
	"github.com/spf13/viper"
)

var keysFile = flag.String("wallet-file", "", "wallet file")
var port = flag.Int("port", 6666, "listening port")

func main() {

	flag.Parse()

	walletFile := viper.GetString("vault-file")
	if _, err := os.Stat(walletFile); err != nil {
		log.Fatalf("Error: wallet file %q missing, ", walletFile)
	}

	vault, err := eosvault.NewVaultFromWalletFile(walletFile)
	if err != nil {
		log.Fatalf("Error: loading vault, %s", err)
	}

	boxer, err := eosvault.SecretBoxerForType(vault.SecretBoxWrap)
	if err != nil {
		log.Fatalf("secret boxer, %s", err)
	}

	vault.Open(boxer)

	keyBag := vault.KeyBag

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
