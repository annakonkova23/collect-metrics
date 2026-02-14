package generator_key

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func Generate(path string) {

	if path == "" {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		path = dir
	}

	privatePath := filepath.Join(path, "private.pem")
	publicPath := filepath.Join(path, "public.pem")

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	privateFile, err := os.Create("private.pem")
	if err != nil {
		panic(err)
	}
	defer privateFile.Close()

	privatePEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := pem.Encode(privateFile, privatePEM); err != nil {
		panic(err)
	}

	publicFile, err := os.Create("public.pem")
	if err != nil {
		panic(err)
	}
	defer publicFile.Close()

	publicBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}

	publicPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicBytes,
	}
	if err := pem.Encode(publicFile, publicPEM); err != nil {
		panic(err)
	}

	fmt.Printf("Ключи созданы: %s, %s \n", publicPath, privatePath)
}
