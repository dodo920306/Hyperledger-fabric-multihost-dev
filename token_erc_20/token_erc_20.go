/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"
	"strings"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

/*Origin*/
// const (
// 	mspID        = "Org1MSP"
// 	cryptoPath   = "../../test-network/organizations/peerOrganizations/org1.example.com"
// 	certPath     = cryptoPath + "/users/User1@org1.example.com/msp/signcerts/cert.pem"
// 	keyPath      = cryptoPath + "/users/User1@org1.example.com/msp/keystore/"
// 	tlsCertPath  = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
// 	peerEndpoint = "localhost:7051"
// 	gatewayPeer  = "peer0.org1.example.com"
// )

var (
	mspID        = "BISCEOrderer1MSP"
	cryptoPath   = "/etc/hyperledger/users/admin/msp"
	certPath     = cryptoPath + "/signcerts/cert.pem"
	keyPath      = cryptoPath + "/keystore/"
	tlsCertPath  = "/etc/hyperledger/peers/peer0/tls/ca.crt"
	peerEndpoint = "peer0:7051"
	tlsCertPathSplit = strings.Split(tlsCertPath, "/")
	gatewayPeer  = tlsCertPathSplit[len(tlsCertPathSplit) - 3]
)

func main() {

	args := os.Args[1:]
        if len(args) < 1 {
                panic(fmt.Errorf("Usage: token_erc_20 <command> [args...]"))
        }
	// The gRPC client connection should be shared by all Gateway connections to this endpoint
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gw.Close()

	// Override default values for chaincode and channel name as they may differ in testing contexts.
	chaincodeName := "bisce"
	if ccname := os.Getenv("CHAINCODE_NAME"); ccname != "" {
		chaincodeName = ccname
	}

	channelName := "biscechannel1"
	if cname := os.Getenv("CHANNEL_NAME"); cname != "" {
		channelName = cname
	}

	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	funcMap := map[string]func(*client.Contract, []string){
		"mint": mint,
		"burn": burn,
		"transfer": transfer,
		"balanceOf": balanceOf,
		"clientAccountBalance": clientAccountBalance,
		"clientAccountID": clientAccountID,
		"totalSupply": totalSupply,
		"approve": approve,
		"allowance": allowance,
		"transferFrom": transferFrom,
		"name": name,
		"symbol": symbol,
		"decimals": decimals,
		"help": help,
	}

    if f, ok := funcMap[args[0]]; ok {
		f(contract, args[1:])
	} else {
		panic(fmt.Errorf("Invalid command: %s", args[0]))
	}
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	files, err := os.ReadDir(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key directory: %w", err))
	}
	privateKeyPEM, err := os.ReadFile(path.Join(keyPath, files[0].Name()))

	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}

func mint(contract *client.Contract, args []string) {
    if len(args) != 1 {
        panic(fmt.Errorf("Usage: mint <amount>"))
    }

	fmt.Println("--> Submit Transaction: Mint.")

	_, err := contract.SubmitTransaction("Mint", args[0])
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Println("*** Transaction committed successfully")
}

func burn(contract *client.Contract, args []string) {
    if len(args) != 1 {
        panic(fmt.Errorf("Usage: burn <amount>"))
    }

	fmt.Println("--> Submit Transaction: Burn.")

	_, err := contract.SubmitTransaction("Burn", args[0])
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Println("*** Transaction committed successfully")
}

func transfer(contract *client.Contract, args []string) {
	if len(args) != 2 {
        panic(fmt.Errorf("Usage: transfer <recipient> <amount>"))
    }

	fmt.Println("--> Submit Transaction: Transfer.")

	_, err := contract.SubmitTransaction("Transfer", args[0], args[1])
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Println("*** Transaction committed successfully")
}

func balanceOf(contract *client.Contract, args []string) {
    if len(args) != 1 {
        panic(fmt.Errorf("Usage: balanceOf <account>"))
    }

	fmt.Println("--> Evaluate Transaction: BalanceOf.")

	evaluateResult, err := contract.EvaluateTransaction("BalanceOf", args[0])
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result: %s\n", result)
}

func clientAccountBalance(contract *client.Contract, args []string) {
    if len(args) != 0 {
        panic(fmt.Errorf("Usage: clientAccountBalance"))
    }

	fmt.Println("--> Evaluate Transaction: ClientAccountBalance. ")

	evaluateResult, err := contract.EvaluateTransaction("ClientAccountBalance")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result: %s\n", result)
}

func clientAccountID(contract *client.Contract, args []string) {
    if len(args) != 0 {
        panic(fmt.Errorf("Usage: clientAccountID"))
    }

	fmt.Println("--> Evaluate Transaction: ClientAccountID. ")

	evaluateResult, err := contract.EvaluateTransaction("ClientAccountID")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	fmt.Printf("*** Result: %s\n", evaluateResult)
}

func totalSupply(contract *client.Contract, args []string) {
	if len(args) != 0 {
        panic(fmt.Errorf("Usage: totalSupply"))
    }

	fmt.Println("--> Evaluate Transaction: TotalSupply. ")

	evaluateResult, err := contract.EvaluateTransaction("TotalSupply")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result: %s\n", result)
}

func approve(contract *client.Contract, args []string) {
	if len(args) != 2 {
        panic(fmt.Errorf("Usage: approve <spender> <value>"))
    }

	fmt.Println("--> Submit Transaction: Approve.")

	_, err := contract.SubmitTransaction("Approve", args[0], args[1])
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Println("*** Transaction committed successfully")
}

func allowance(contract *client.Contract, args []string) {
	if len(args) != 2 {
        panic(fmt.Errorf("Usage: approve <owner> <spender>"))
    }

	fmt.Println("--> Evaluate Transaction: Allowance. ")

	evaluateResult, err := contract.EvaluateTransaction("Allowance", args[0], args[1])
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("*** Result: %s\n", result)
}

func transferFrom(contract *client.Contract, args []string) {
	if len(args) != 3 {
		panic(fmt.Errorf("Usage: transferFrom <from> <to> <value>"))
	}

	fmt.Println("--> Submit Transaction: TransferFrom. ")

	_, err := contract.SubmitTransaction("TransferFrom", args[0], args[1], args[2])
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Println("*** Transaction committed successfully")
}

func name(contract *client.Contract, args []string) {
	if len(args) != 0 {
        panic(fmt.Errorf("Usage: name"))
    }

	fmt.Println("--> Evaluate Transaction: Name. ")

	evaluateResult, err := contract.EvaluateTransaction("Name")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	fmt.Printf("*** Result: %s\n", string(evaluateResult))
}

func symbol(contract *client.Contract, args []string) {
	if len(args) != 0 {
        panic(fmt.Errorf("Usage: symbol"))
    }

	fmt.Println("--> Evaluate Transaction: Symbol. ")

	evaluateResult, err := contract.EvaluateTransaction("Symbol")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	fmt.Printf("*** Result: %s\n", string(evaluateResult))
}

func decimals(contract *client.Contract, args []string) {
        if len(args) != 0 {
        panic(fmt.Errorf("Usage: decimals"))
    }

        fmt.Println("--> Evaluate Transaction: Decimals. ")

        evaluateResult, err := contract.EvaluateTransaction("Decimals")
        if err != nil {
                panic(fmt.Errorf("failed to evaluate transaction: %w", err))
        }

        fmt.Printf("*** Result: %s\n", string(evaluateResult))
}

func help(contract *client.Contract, args []string) {
    fmt.Print(`Usage:
  token_erc_20 command [arguments]
Commands:
  mint                  <amount>                | Creates new tokens and adds them to minter's account balance.
  burn                  <amount>                | Redeems tokens the minter's account balance.
  transfer              <recipient> <amount>    | Transfers tokens from client account to recipient account.
  balanceOf             <account>               | Returns the balance of the given account.
  clientAccountBalance                          | Returns the balance of the requesting client's account.
  clientAccountID                               | Returns the id of the requesting client's account.
  totalSupply                                   | Returns the total token supply.
  approve               <spender> <value>       | Allows the spender to withdraw from the calling client's token account.
  allowance             <owner>	<spender>       | Returns the amount still available for the spender to withdraw from the owner.
  transferFrom          <from> <to> <value>     | Transfers the value amount from the \"from\" address to the \"to\" address.
  name                                          | Returns a descriptive name for fungible tokens in this contract.
  symbol                                        | Returns an abbreviated name for fungible tokens in this contract.
  decimals                                      | Returns the decimals for fungible tokens in this contract.
  help                                          | Show summary
`)
}

// Format JSON data
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}
