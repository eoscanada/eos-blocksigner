package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
	"net/http"
	"time"

	"os"

	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	eosvault "github.com/eoscanada/eosc/vault"
)

var keysFile = flag.String("keys-file", "", "keys file")
var walletFile = flag.String("wallet-file", "", "wallet file")
var kmsGCPKeypath = flag.String("kms-gcp-keypath", "", "cryptoKeys path to GCP's KMS system")
var port = flag.Int("port", 6666, "listening port")

var zlog = zap.NewNop()


func main() {
	flag.Parse()

	SetupLogger()

	if *keysFile != "" && *walletFile != "" {
		zlog.Fatal("Error: --keys-file and --wallet-file should not be use together")
		os.Exit(2)
	}

	if *keysFile == "" && *walletFile == "" {
		zlog.Fatal("Error: Require one of flags --keys-file and --wallet-file")
		os.Exit(3)
	}

	var keyBag *eos.KeyBag
	if *walletFile != "" {
		if _, err := os.Stat(*walletFile); err != nil {
			zlog.Fatal("Error: wallet file %q missing", zap.String("walletFile", *walletFile))
			os.Exit(4)
		}

		vault, err := eosvault.NewVaultFromWalletFile(*walletFile)
		if err != nil {
			zlog.Fatal("Error: loading vault", zap.Error(err))
			os.Exit(5)
		}

		boxer, err := eosvault.SecretBoxerForType(vault.SecretBoxWrap, *kmsGCPKeypath)
		if err != nil {
			zlog.Fatal("Error: secret boxer", zap.Error(err))
			os.Exit(6)
		}

		if err := vault.Open(boxer); err != nil {
			zlog.Fatal("Error: open vault", zap.Error(err))
			os.Exit(7)
		}

		keyBag = vault.KeyBag
	}

	if *keysFile != "" {
		keyBag = eos.NewKeyBag()

		if err := keyBag.ImportFromFile(*keysFile); err != nil {
			zlog.Fatal("Error: import keys from file", zap.Error(err))
			os.Exit(8)
		}
	}

	http.HandleFunc("/v1/wallet/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	http.HandleFunc("/v1/wallet/get_public_keys", func(w http.ResponseWriter, r *http.Request) {
		zlog.Info("Handling get_public_keys")

		var out []string
		for _, key := range keyBag.Keys {
			out = append(out, key.PublicKey().String())
		}

		_ = json.NewEncoder(w).Encode(out)
	})

	http.HandleFunc("/v1/wallet/sign_digest", func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()

		var inputs []string
		if err := json.NewDecoder(r.Body).Decode(&inputs); err != nil {
			zlog.Error("Error: sign_digest", zap.Error(err))
			http.Error(w, "couldn't decode input params", 500)
			return
		}

		digest, err := hex.DecodeString(inputs[0])
		if err != nil {
			zlog.Error("Error: digest decode", zap.Error(err))
			http.Error(w, "couldn't decode digest", 500)
		}

		pubKey, err := ecc.NewPublicKey(inputs[1])
		if err != nil {
			zlog.Error("Error: public key", zap.Error(err))
			http.Error(w, "couldn't decode public key", 500)
		}

		signed, err := keyBag.SignDigest(digest, pubKey)
		if err != nil {
			zlog.Error("Error: signing", zap.Error(err))
			http.Error(w, "signing error", 500)
			return
		}

		zlog.Info("Signing digest", zap.String("digest", hex.EncodeToString(digest)), zap.String("pubKey", pubKey.String()))

		w.WriteHeader(201)
		err = json.NewEncoder(w).Encode(signed)
		elapsedTime := time.Since(t0)
		if err != nil {
			zlog.Error("Error: encoding", zap.Error(err), zap.Duration("elapsedTime", elapsedTime))
		} else {
			zlog.Info("done", zap.Duration("elapsedTime", elapsedTime))
		}
	})

	address := "127.0.0.1"
	listeningOn := fmt.Sprintf("%s:%d", address, *port)
	zlog.Info("Listening for block signing operations", zap.String("address", address), zap.Int("port", *port))
	if err := http.ListenAndServe(listeningOn, nil); err != nil {
		fmt.Printf("Failed listening on port %s: %s\n", listeningOn, err)
		zlog.Fatal("Failed listening", zap.Error(err), zap.String("listeningOn", listeningOn))
	}
}


func errorCheck(prefix string, err error) {
	if err != nil {
		zlog.Fatal(prefix, zap.Error(err))
		os.Exit(1)
	}
}


func SetupLogger() {
	var err error
	if _, err = os.Stat("/.dockerenv"); !os.IsNotExist(err) {
		zlog, err = zapdriver.NewProduction()
	} else {
		zlog, err = zap.NewDevelopment()
	}
	errorCheck("setting up zap logger", err)
}