package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type AuctionContract struct {
	contractapi.Contract
}

// type AuctionTransactionContext struct {
// 	contractapi.TransactionContext
// 	data []byte
// }

// type AuctionTransactionContextInterface interface {
// 	contractapi.TransactionContextInterface
// 	StartAuction(string, string, int) error
// }

func (c *AuctionContract) AuctionExists(ctx contractapi.TransactionContextInterface, auctionID string) (bool, error) {
	data, err := ctx.GetStub().GetState(auctionID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

func (c *AuctionContract) StartAuction(ctx contractapi.TransactionContextInterface, auctionID string, item string, startingPrice uint32) error {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("Failed reading from world state: %s", err)
	} else if exists {
		return fmt.Errorf("Auction with ID = %s already exists", auctionID)
	}

	clientID, err := c.GetClientId(ctx)
	//clientID := "abcd123"
	if err != nil {
		return fmt.Errorf("Getting cliend ID failed: %s", err)
	}

	bids := make(map[string]uint32)

	auction := new(Auction)
	auction.Type = "auction"
	auction.Item = item
	auction.Seller = clientID
	auction.Bids = bids
	auction.Winner = ""
	auction.Price = startingPrice
	auction.Status = "active"

	bytes, _ := json.Marshal(auction)

	//eventPayload := new()

	return ctx.GetStub().PutState(auctionID, bytes)
}

func (c *AuctionContract) GetAuction(ctx contractapi.TransactionContextInterface, auctionID string) (*Auction, error) {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return nil, fmt.Errorf("Failed reading from world state: %s", err)
	} else if !exists {
		return nil, fmt.Errorf("Auction with ID = %s does not exist", auctionID)
	}

	bytes, _ := ctx.GetStub().GetState(auctionID)

	// cannot be trusted
	// does not work either
	// possibly would work from an application level
	// ts, tserr := shim.ChaincodeStubInterface(ctx.GetStub()).GetTxTimestamp()

	// if tserr != nil {
	//  	return nil, fmt.Errorf("Getting transaction timestamp failed: %s %s", ts, tserr)
	// }
	// fmt.Printf("Transaction timestamp: %s", ts)

	auction := new(Auction)

	err = json.Unmarshal(bytes, auction)

	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal world state data to type Auction")
	}

	return auction, nil
}

func (c *AuctionContract) CloseAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
	//get auction and check if exists
	bytes, err := ctx.GetStub().GetState(auctionID)

	if err != nil {
		return err
	}

	auction := new(Auction)
	err = json.Unmarshal(bytes, auction)
	if err != nil {
		return fmt.Errorf("Could not unmarshal world state data to type Auction")
	}

	//get clientID and make sure it matches the seller field of the auction
	clientID, err := c.GetClientId(ctx)
	if err != nil {
		return fmt.Errorf("Failed getting client identity %v", err)
	}

	if clientID != auction.Seller {
		return fmt.Errorf("Auction can only be ended by the client that started it")
	}

	if auction.Status != "active" {
		return fmt.Errorf("Auction can only be ended if it is in an \"active\" state")
	}

	//TODO
	//handle draws
	maxOffer := auction.Price
	winner := auction.Seller
	draw := false
	for user, offer := range auction.Bids {
		if offer > maxOffer {
			maxOffer = offer
			winner = user
			draw = false
		}
		if offer == maxOffer {
			draw = true
		}
	}
	if draw {
		return fmt.Errorf("Cannot close the auction: draw")
	}
	//TODO: handle somehow?
	if winner == auction.Seller {
		return fmt.Errorf("Cannot close the auction: no winning bids")
	}

	auction.Status = string("closed")
	auction.Winner = winner
	auction.Price = maxOffer
	closedAuction, _ := json.Marshal(auction)
	err = ctx.GetStub().PutState(auctionID, closedAuction)
	if err != nil {
		return fmt.Errorf("Failed closing the auction %v", err)
	}

	//successfully closed the auction
	return nil
}

func (c *AuctionContract) UpdateAuction(ctx contractapi.TransactionContextInterface, auctionID string, newItem string) error {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The asset %s does not exist", auctionID)
	}

	auction := new(Auction)
	auction.Type = "auction"
	auction.Item = newItem

	bytes, _ := json.Marshal(auction)

	return ctx.GetStub().PutState(auctionID, bytes)
}

func (c *AuctionContract) DeleteAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("Reading from world state failed: %s", err)
	} else if !exists {
		return fmt.Errorf("The auction with ID = %s does not exist", auctionID)
	}

	return ctx.GetStub().DelState(auctionID)
}

func (c *AuctionContract) Bid(ctx contractapi.TransactionContextInterface, auctionID string, offer uint32) error {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("Reading from world state failed: %s", err)
	} else if !exists {
		return fmt.Errorf("The auction with ID = %s does not exist", auctionID)
	}

	//get auction
	bytes, err := ctx.GetStub().GetState(auctionID)

	if err != nil {
		return err
	}

	auction := new(Auction)
	err = json.Unmarshal(bytes, auction)
	if err != nil {
		return fmt.Errorf("Could not unmarshal world state data to type Auction")
	}

	if auction.Status != "active" {
		return fmt.Errorf("The auction is not active")
	}

	clientID, err := c.GetClientId(ctx)
	if err != nil {
		return fmt.Errorf("Failed getting client identity %v", err)
	}

	previousBid := auction.Bids[clientID]

	if offer <= previousBid {
		return fmt.Errorf("User already placed a higher bid before")
	}

	// auction.Price is not changed during auction now, so it stays the starting price
	// checking if higher than current hihgest bid would help the bad actors - but also reduce the amount of transactions
	if offer <= uint32(auction.Price) {
		return fmt.Errorf("Offer lower than the starting price")
	}

	auction.Bids[clientID] = offer
	bytes, _ = json.Marshal(auction)

	//eventPayload := new()

	return ctx.GetStub().PutState(auctionID, bytes)
}
