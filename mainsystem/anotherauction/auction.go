package main

//TODO:
// finish date
// no network-wide time, timestamp is put into transaction by the submitting client
// also: timestamppb is crashing my application (go builder failed, yet go build works)

// transactions based on time are very problematic as they are non-deterministic and break consensus
// a paper on this: https://uwspace.uwaterloo.ca/handle/10012/16784
// short explanation by the authors on youtube (also intro on permissioned vs permissionless):
// TimeFabric https://www.youtube.com/watch?v=3H7adN8JISg
// but it is implemented on top of Fabric 1.4 (2021) which has bigger issues than lack of time

type Auction struct {
	// every object needs a type field - they are referenced by ID only
	// and we can never know what ctx.GetStub().GetState(auctionID) will return
	Type   string            `json:"objectType"`
	Item   string            `json:"item"`
	Seller string            `json:"seller"`
	Bids   map[string]uint32 `json:"bids"`
	Winner string            `json:"winner"`
	Price  uint32            `json:"price"`
	//enums are not supported by fabric so we have to manually emulate them
	Status string `json:"status"`
}

//TODO:
//list of active auctions
type ActiveAuctions struct {
	Type string `json:"objectType"`
}

// I have decided just storing the price should be enough for now

// type Bid struct {
// 	Type   string `json:"objectType"`
// 	Price  uint32 `json:"price"`
// 	Bidder string `json:"bidder"`
// }
