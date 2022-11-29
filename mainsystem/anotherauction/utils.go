package main

import (
	"encoding/base64"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric/protoutil"
)

func (c *AuctionContract) GetClientId(ctx contractapi.TransactionContextInterface) (string, error) {

	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Reading clientID failed: %s", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("Decoding base64 clientID failed: %s", err)
	}
	return string(decodeID), nil
}

//from Hyperledger Fabric's documentation:
// The epoch in which this header was generated, where epoch is defined based on block height
// Epoch in which the response has been generated. This field identifies a
// logical window of time. A proposal response is accepted by a peer only if
// two conditions hold:
// 1. the epoch specified in the message is the current epoch
// 2. this message has been only seen once during this epoch (i.e. it hasn't
//    been replayed)
//Epoch uint64 `protobuf:"varint,6,opt,name=epoch,proto3" json:"epoch,omitempty"`

// when looking for references to epoch in the source code I have found TODO comments
// saying that it still needs to be implemented and it currently returns 0 every time
func (c *AuctionContract) GetEpoch(ctx contractapi.TransactionContextInterface) (uint64, error) {
	proposal, err := ctx.GetStub().GetSignedProposal()

	if err == nil {
		bytes := proposal.GetProposalBytes()
		proposal, err := protoutil.UnmarshalProposal(bytes)
		if err == nil {
			if proposal.GetHeader() != nil {
				header, err := protoutil.UnmarshalHeader(proposal.GetHeader())
				if err == nil {
					channelHeader := protoutil.UnmarshalChannelHeaderOrPanic(header.GetChannelHeader())
					return channelHeader.Epoch, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("Get epoch failed")
}
