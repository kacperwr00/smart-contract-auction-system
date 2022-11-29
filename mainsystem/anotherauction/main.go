package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
)

//cutting a block every transaction (BatchSize of MaxMessageCount:1)?

func main() {
	auctionContract := new(AuctionContract)
	auctionContract.Info.Version = "0.0.1"
	auctionContract.Info.Description = "Auction system Smart Contracts"
	auctionContract.Info.License = new(metadata.LicenseMetadata)
	auctionContract.Info.License.Name = "Apache-2.0"
	auctionContract.Info.Contact = new(metadata.ContactMetadata)
	auctionContract.Info.Contact.Name = "Kacper Wróblewski"

	chaincode, err := contractapi.NewChaincode(auctionContract)
	chaincode.Info.Title = "Auction system chaincode"
	chaincode.Info.Version = "0.0.1"

	if err != nil {
		panic("Could not create chaincode from auctionContract." + err.Error())
	}

	err = chaincode.Start()

	if err != nil {
		panic("Failed to start chaincode. " + err.Error())
	}
}
