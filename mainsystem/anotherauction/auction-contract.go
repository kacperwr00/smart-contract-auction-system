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

// func (AuctionContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

// 	funcName, params := stub.GetFunctionAndParameters()

// 	stub.GetTransient()

// 	indexName := "txID~key"

// 	if funcName == "addNewKey" {

// 		key := params[0]
// 		value := params[1]

// 		keyTxIdKey, err := stub.CreateCompositeKey(indexName, []string{stub.GetTxID(), key})
// 		if err != nil {
// 			return shim.Error(err.Error())
// 		}

// 		creator, _ := stub.GetCreator()

// 		// Add key and value to the state
// 		stub.PutState(key, []byte(value))
// 		stub.PutState(keyTxIdKey, creator)

// 	} else if funcName == "checkTxID" {
// 		txID := params[0]

// 		it, _ := stub.GetStateByPartialCompositeKey(indexName, []string{txID})

// 		for it.HasNext() {
// 			keyTxIdRange, err := it.Next()
// 			if err != nil {
// 				return shim.Error(err.Error())
// 			}

// 			_, keyParts, _ := stub.SplitCompositeKey(keyTxIdRange.Key)
// 			key := keyParts[1]
// 			fmt.Printf("key affected by txID %s is %s\n", txID, key)
// 			txIDCreator := keyTxIdRange.Value

// 			sId := &msp.SerializedIdentity{}
// 			err := proto.Unmarshal(txIDCreator, sId)
// 			if err != nil {
// 				return shim.Error(fmt.Sprintf("Could not deserialize a SerializedIdentity, err %s", err))
// 			}

// 			bl, _ := pem.Decode(sId.IdBytes)
// 			if bl == nil {
// 				return shim.Error(fmt.Sprintf("Could not decode the PEM structure"))
// 			}
// 			cert, err := x509.ParseCertificate(bl.Bytes)
// 			if err != nil {
// 				return shim.Error(fmt.Sprintf("ParseCertificate failed %s", err))
// 			}

// 			fmt.Printf("Certificate of txID %s creator is %s", txID, cert)
// 		}
// 	}

// 	return shim.Success(nil)
// }

func (c *AuctionContract) AuctionExists(ctx contractapi.TransactionContextInterface, auctionID string) (bool, error) {
	data, err := ctx.GetStub().GetState(auctionID)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

func (c *AuctionContract) StartAuction(ctx contractapi.TransactionContextInterface, auctionID string, item string, startingPrice uint32, duration uint32) error {
	if len(auctionID) < 6 {
		return fmt.Errorf("Auction name has to be at least 6 charachters long")
	}
	//because active is used as a prefix when creating the index allowing to query all active auctions
	if auctionID[:5] == "active" {
		return fmt.Errorf("Auction's name can not start with active")
	}

	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("Failed reading from world state: %s", err)
	} else if exists {
		return fmt.Errorf("Auction with ID = %s already exists", auctionID)
	}

	clientID, err := c.GetClientId(ctx)
	if err != nil {
		return fmt.Errorf("Getting cliend ID failed: %s", err)
	}

	ts, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("Gettimestamp failed, err: %s", err)
	}

	bids := make(map[string]Bid)

	auction := new(Auction)
	auction.Type = "auction"
	auction.Item = item
	auction.Seller = clientID
	auction.Bids = bids
	auction.Winner = ""
	auction.Price = startingPrice
	auction.Status = "active"
	auction.FinishTS = ts.Seconds + int64(duration)

	bytes, _ := json.Marshal(auction)

	indexState := "AuctionState~AuctionID"
	auctionStateIndexKey, err := ctx.GetStub().CreateCompositeKey(indexState, []string{auction.Status, auctionID})

	err = ctx.GetStub().PutState(auctionStateIndexKey, []byte(clientID))
	if err != nil {
		return fmt.Errorf("Creating the state index failed, err: %s", err)
	}

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

	// only allow to close the auction if the specified duration passed
	ts, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("Could not obtain the transaction timestamp, %v", err)
	}
	if ts.Seconds < auction.FinishTS {
		return fmt.Errorf("Could not close the auction, the specified time duration did not pass yet")
	}

	//get clientID and make sure it matches the seller field of the auction or a bidder
	clientID, err := c.GetClientId(ctx)
	if err != nil {
		return fmt.Errorf("Failed getting client identity %v", err)
	}

	if _, clientIsABidder := auction.Bids[clientID]; !clientIsABidder && clientID != auction.Seller {
		return fmt.Errorf("Auction can only be ended by the client that started it")
	}

	if auction.Status != "active" {
		return fmt.Errorf("Auction can only be ended if it is in an \"active\" state")
	}

	// find the winning bid
	// Warining: if there were two exact same bids in the exact same nanosecond the one that will be checked first will win
	maxOffer := auction.Price
	// uncomment for second price auction
	// finishPrice := maxOffer
	maxSec := auction.FinishTS
	maxNano := int32(0)
	winner := auction.Seller
	for user, bid := range auction.Bids {
		if bid.TSSec < auction.FinishTS {
			if bid.Price > maxOffer ||
				(bid.Price == maxOffer &&
					(bid.TSSec < maxSec ||
						(bid.TSSec == maxSec && bid.TSNano < maxNano))) {
				// finishPrice = maxOffer
				maxOffer = bid.Price
				winner = user
				maxSec = bid.TSSec
				maxNano = bid.TSNano
			} //else if bid.Price > finishPrice {
			//	finishPrice = bid.Price
			//}
		}
	}

	indexState := "AuctionState~AuctionID"
	auctionStateIndexKey, err := ctx.GetStub().CreateCompositeKey(indexState, []string{auction.Status, auctionID})

	err = ctx.GetStub().DelState(auctionStateIndexKey)

	if err != nil {
		return fmt.Errorf("Failed deleting entry from the active index, %v", err)
	}

	auction.Status = "closed"
	if winner == auction.Seller {
		auction.Status = "noBidsReceived"
	}
	auction.Winner = winner
	auction.Price = maxOffer
	// auction.Price = finishPrice
	closedAuction, _ := json.Marshal(auction)
	err = ctx.GetStub().PutState(auctionID, closedAuction)
	if err != nil {
		return fmt.Errorf("Failed closing the auction %v", err)
	}

	//successfully closed the auction
	return nil
}

// func (c *AuctionContract) DeleteAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
// 	exists, err := c.AuctionExists(ctx, auctionID)
// 	if err != nil {
// 		return fmt.Errorf("Reading from world state failed: %s", err)
// 	} else if !exists {
// 		return fmt.Errorf("The auction with ID = %s does not exist", auctionID)
// 	}
//
// 	return ctx.GetStub().DelState(auctionID)
// }

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
	if clientID == auction.Seller {
		return fmt.Errorf("Cannot bid on one's own auctions")
	}

	previousBid := auction.Bids[clientID]

	if offer <= previousBid.Price {
		return fmt.Errorf("User already placed a higher bid before")
	}

	// auction.Price is not changed during auction now, so it stays the starting price
	// checking if higher than current hihgest bid would help the bad actors,
	// by disallowing anyone else to bid on an auction by placing an arbitrarily large bid
	// that they would not intend on keeping - but also reduce the amount of transactions
	if offer <= uint32(auction.Price) {
		return fmt.Errorf("Offer lower than the starting price")
	}

	ts, err := ctx.GetStub().GetTxTimestamp()

	if err != nil {
		return fmt.Errorf("Could not obtain transaction timestamp, %v", err)
	}
	if ts.Seconds >= int64(auction.FinishTS) {
		return fmt.Errorf("Too late to bid on this auction")
	}

	bid := new(Bid)
	bid.Price = offer
	bid.Type = "bid"
	bid.TSSec = ts.Seconds
	bid.TSNano = ts.Nanos

	auction.Bids[clientID] = *bid
	bytes, _ = json.Marshal(auction)

	//eventPayload := new()

	return ctx.GetStub().PutState(auctionID, bytes)
}

func (c *AuctionContract) GetActiveAuctions(ctx contractapi.TransactionContextInterface) ([]string, error) {
	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey("AuctionState~AuctionID", []string{"active"})
	if err != nil {
		return nil, fmt.Errorf("Error occured while getting active auctions: %s", err)
	}
	res := []string{}
	if !iterator.HasNext() {
		return res, nil
	}

	for iterator.HasNext() {
		compositeAuctionID, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("Failed creating the active auction list, %v", err)
		}

		_, parts, err := ctx.GetStub().SplitCompositeKey(compositeAuctionID.Key)
		if err != nil {
			return nil, fmt.Errorf("Failed creating the active auction list, %v", err)
		}
		res = append(res, parts[1])
	}

	return res, nil
}

// func (c *AuctionContract) getHistoryForUFO(stub shim.ChaincodeStubInterface) (string, error) {
// 	params := []string{"get", "name"}
// 	queryArgs := make([][]byte, len(params))
// 	for i, arg := range params {
// 		queryArgs[i] = []byte(arg)
// 	}

// 	response := stub.InvokeChaincode("sacc", queryArgs, "mychannel")
// 	if response.Status != shim.OK {
// 		return "", fmt.Errorf("Failed to query chaincode. Got error: %s", response.Payload)
// 	}
// 	return string(response.Payload), nil
// }

// func (mc *AuctionContract) Testing(ctx contractapi.TransactionContextInterface) error {
// 	stub := ctx.GetStub()

// 	params := []string{"GetChainInfo", stub.GetChannelID()}
// 	invokeArgs := make([][]byte, len(params))
// 	for i, arg := range params {
// 		invokeArgs[i] = []byte(arg)
// 	}

// 	//"If `channel` is empty, the caller's channel is assumed."
// 	resp := stub.InvokeChaincode("qscc", invokeArgs, stub.GetChannelID())

// 	if resp.Status != shim.OK {
// 		return fmt.Errorf("Failed to query chaincode. Got error: %s", resp.Message)
// 	}
// 	return fmt.Errorf("Succesfully queried chaincode. Got payload: %s", resp.Payload)
// }

// func (mc *AuctionContract) Testing2(ctx contractapi.TransactionContextInterface) error {
// 	epoch, err := mc.GetEpoch(ctx)
// 	if err == nil {
// 		return fmt.Errorf("GetEpoch successful - %d", epoch)
// 	}
// 	return fmt.Errorf("GetEpoch failed, err: %s", err)
// }

// func (mc *AuctionContract) Testing3(ctx contractapi.TransactionContextInterface) error {
// 	ts, err := ctx.GetStub().GetTxTimestamp()
// 	if err == nil {
// 		return fmt.Errorf("GetTimestamp successful - %s", ts)
// 	}
// 	return fmt.Errorf("Gettimestamp failed, err: %s", err)

// 	//ctx.GetStub().GetHistoryForKey()
// }
