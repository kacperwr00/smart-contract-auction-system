package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
)

func main() {
	auctionContract := new(AuctionContract)
	auctionContract.Info.Version = "0.0.1"
	auctionContract.Info.Description = "My Smart Contract"
	auctionContract.Info.License = new(metadata.LicenseMetadata)
	auctionContract.Info.License.Name = "Apache-2.0"
	auctionContract.Info.Contact = new(metadata.ContactMetadata)
	auctionContract.Info.Contact.Name = "John Doe"

	chaincode, err := contractapi.NewChaincode(auctionContract)
	chaincode.Info.Title = "myproject chaincode"
	chaincode.Info.Version = "0.0.1"

	if err != nil {
		panic("Could not create chaincode from MyAuctionContract." + err.Error())
	}

	err = chaincode.Start()

	if err != nil {
		panic("Failed to start chaincode. " + err.Error())
	}
}
