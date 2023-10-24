package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/open-policy-agent/opa/rego"
)

func main() {

	err := genToken()
	if err != nil {
		log.Fatal(err)
	}

	err = genRsaKey()
	if err != nil {
		log.Fatal(err)
	}

	err = generateECDSAKeys()
	if err != nil {
		log.Fatal(err)
	}
}

func genToken() error {

	claims := struct {
		jwt.RegisteredClaims
		Roles []string `json:"roles"`
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "service project",
			Subject:   "12345678",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Roles: []string{"ADMIN"},
	}

	method := jwt.GetSigningMethod(jwt.SigningMethodRS256.Name)
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	file, err := os.Open("zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem")
	if err != nil {
		return fmt.Errorf("opening private file: %w", err)
	}
	defer file.Close()

	pemData, err := io.ReadAll(io.LimitReader(file, 1024*1024))
	if err != nil {
		return fmt.Errorf("reading auth private file: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemData)
	if err != nil {
		return fmt.Errorf("parsing private key: %w", err)
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return fmt.Errorf("generating key: %w", err)
	}

	fmt.Println("********** TOKEN **********")
	fmt.Println(str)
	fmt.Println("********** PUBILC **********")

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the public key file.
	if err := pem.Encode(os.Stdout, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))
	var clm struct {
		jwt.RegisteredClaims
		Roles []string
	}

	kf := func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	}
	tkn, err := parser.ParseWithClaims(str, &clm, kf)
	if err != nil {
		return err
	}

	if !tkn.Valid {
		return fmt.Errorf("token is invalid")
	}

	fmt.Println("Token is valid")

	// ----------------------------------------
	// Construct a PEM block for the private key.
	var b bytes.Buffer
	// Write the private key to the private key file.
	if err := pem.Encode(&b, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w\n", err)
	}

	if err := opaPolicyEvaluationAuthen(context.Background(), b.String(), str, clm.Issuer); err != nil {
		return fmt.Errorf("OPS authentication failed: %w\n", err)
	}
	fmt.Println("token validated by opa")

	// ----------------------------------------

	if err := opaPolicyEvaluationAuthor(context.Background()); err != nil {
		return fmt.Errorf("OPS authorization failed: %w\n", err)
	}
	fmt.Println("auth validated by opa")

	fmt.Printf("%#v\n", clm)
	return nil
}

// opaPolicyEvaluation asks opa to evaulate the token against the specified token
// policy and public key.

// Core OPA policies.
var (
	//go:embed rego/authentication.rego
	opaAuthentication string
	//go:embed rego/authorization.rego
	opaAuthorization string
)

func opaPolicyEvaluationAuthor(ctx context.Context) error {
	rule := "ruleAdminOnly"
	opaPackage := "ardan.rego"
	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)

	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", opaAuthorization),
	).PrepareForEval(context.Background())
	if err != nil {
		return err
	}

	input := map[string]interface{}{
		"Roles":   []string{"ADMIN"},
		"Subject": "1234567",
		"UserID":  "1234567",
	}

	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings results[%v] ok[%v]", results, ok)
	}

	return nil
}

func opaPolicyEvaluationAuthen(ctx context.Context, pem string, tokenString string, issuer string) error {
	rule := "auth"
	opaPackage := "ardan.rego"
	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)

	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", opaAuthentication),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	input := map[string]interface{}{
		"Key":   pem,
		"Token": tokenString,
		"ISS":   issuer,
	}
	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings results[%v] ok[%v]", results, ok)
	}

	return nil
}

func genRsaKey() error {
	// Generate a new private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generating key: %w", err)
	}

	// Create a file for the private key information in PEM form.
	privateFile, err := os.Create("rsa-private.pem")
	if err != nil {
		return fmt.Errorf("creating private file: %w", err)
	}
	defer privateFile.Close()

	// Construct a PEM block for the private key.
	privateBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Write the private key to the private key file.
	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return fmt.Errorf("encoding to private file: %w", err)
	}

	// Create a file for the public key information in PEM form.
	publicFile, err := os.Create("rsa-public.pem")
	if err != nil {
		return fmt.Errorf("creating public file: %w", err)
	}
	defer publicFile.Close()

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the public key file.
	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	fmt.Println("private and public key files generated")
	return nil
}

func generateECDSAKeys() error {
	// Generate a new private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	// Marshal the private key to a DER encoded binary
	der, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return err
	}

	// Create a PEM block for the private key
	privateKeyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: der,
	}

	// Create or overwrite the private key file
	privateFile, err := os.Create("ecdsa-private.pem")
	if err != nil {
		return err
	}
	defer privateFile.Close()

	// Write the PEM encoded private key to the file
	err = pem.Encode(privateFile, privateKeyBlock)
	if err != nil {
		return err
	}

	// Marshal the public key to a DER encoded binary
	der, err = x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	// Create a PEM block for the public key
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	}

	// Create or overwrite the public key file
	publicFile, err := os.Create("ecdsa-public.pem")
	if err != nil {
		return err
	}
	defer publicFile.Close()

	// Write the PEM encoded public key to the file
	err = pem.Encode(publicFile, publicKeyBlock)
	return err
}
