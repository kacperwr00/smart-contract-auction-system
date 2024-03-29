# Choosing the stack

This directory should contain mulitple hello world style smart contract applications (possibly with both transaction processors and client interface) developed with different block chains to help me choose the right one for the main application

## Possible approcaches

We could write smart contracts targetting a public, established blockchain and deploy it, like Ethereum, EOS or Solana. Then the client might use the ready to use interfaces to use these smart contracts in order to deploy transactions to the blockchain. This is a much simpler approach - we would not have to worry about consensus, safety or speed - at least not beyond choosing the right blockchain for the application. On the other I would not be able to take credit for these aspects, and they seem more interesting than creating a client or a GUI. These blockchain also enforce fees or use of a testnet.

We could also create our own blockchain - it's actually quite simple to implement a Bitcoin-style blockchain, especially using a good crypto library. Creating an EVM-style blockchain though, that could execute smart contracts would be a much bigger undertaking. A solution that I am partial to right now would be to use a framework providing the ability to create private blockchains, preferably one from the Hyperledger Project.

## Blockchain comparison

Currently considered blockchains and their pros/cons include:
<ul>
    <li>Solana
        <ul>
            <li> ✅ proof of history and proof of stake - makes it easier to agree on the order of the actions on the shared ledger - can be crucial for an auction system; also reduces overhead at the same time ✅ </li>
            <li> ✅ Transaction fees on Solana are estimated at $10 for 1 million transaction (vs up to $10 per transaction on current Etheruem) ✅ </li>
            <li> ✅ 65,000 transactions per second (just as much as Visa's and 4000 times more than Etheuruem); 400ms block times ✅ </li>
            <li> ✅ Claims it's scalable and secure ✅ </li>
            <li> ✅ Rust / C / C++ ✅ </li>
            <li> ✅ not EVM compatible, but bridge available ✅ </li>
            <li> ✅ JSON RPC API and many different SDKs built on top of it (which annoys me at the same time - why do we need the transaction payload to be human readable and introduce overhead?) ✅ </li>
            <li> ✅ Great documentation and community ✅ </li>
            <li> ✅ Many open source projects available which may be used as examples ✅ </li>
            <li> 🔴 Naturally only includes a single predefined blockchain which can't be configured by the developer to fit the specific needs of the project. You get what you get (a permissionless PoS + PoH blockchain) 🔴 </li>
        </ul>
    </li>
    <li>EOS
        <ul>
            <li> ✅ It's possible to deploy a smart contract with as few lines of code - simplicity (hello world example smart contract contains 6 lines) ✅ </li>
            <li> ✅ Speed / scalability ✅ </li>
            <li> ✅ C++ ✅ </li>
            <li> ✅ Great wallet/account/auth included ✅ </li>
            <li> ✅ Easy to work with; decent documentation ✅ </li>
            <li> 🔴 dPoS - quite centralized, not as private or safe 🔴 </li>
        </ul>
    </li>
    <li>Hyperledger Sawtooth
        <ul>
            <li> ✅ supports both its own smart contracts and ethereum ones ✅ </li>
            <li> ✅ PBFT, PoET, Raft and devmode consensus algorithms all available and it's possible to change them after a blockchain has been created ✅ </li>
            <li> ✅ Python, Go, Javascript, Rust ✅ </li>
            <li> ✅ Examples, reasonably documented ✅ </li>
            <li> ✅ HTTP/JSON client interface ✅ </li>
            <li> ✅ Based on Transaction batches - which might help build a fair auction system ✅ </li>
            <li> ✅ considered safe ✅ </li>
            <li> ✅ supports custom payload formats ✅ </li>
            <li> 🔴 outdated both in documentation and code; doesn't support new Ubuntu or python versions; causes headaches all day long 🔴 </li>
            <li> 🔴 working with it was honestly dreadful 🔴 </li>
            <li> 🔴 more code / work required compared to deploying smart contracts to an existing chain 🔴 </li>
            <li> 🔴 permissioned networks only (kind of. Either way privacy and openess is not it's strong suit) 🔴 </li>
            <li> 🔴 perforamnce and scalability depend on implementation, but can't rival the likes of solana (yet still, the likes of Etheruem can't rival sawtooth's performance - at least at the time of writing) 🔴 </li>
            <li> 🔴 no PoW or PoS 🔴 </li>
        </ul>
    </li>
    <li> Hyperledger Fabric
        <ul>
            <li> ✅ allows for customization of the distributed ledger ✅ </li>
            <li> ✅ supports both its own smart contracts and ethereum ones ✅ </li>
            <li> ✅ verbose, thourough documentation (a bit messy and outdated though) ✅ </li>
            <li> ✅ multiple samples and examples ✅ </li>
            <li> 20000 transactions per second advertised (depends on configuration and hardware on the network). Again, both Visa and Solana achieve around 70000; performance should still be good enough for permissioned enterprise use cases. </li>
            <li> 🔴 permissioned networks only 🔴 </li>
            <li> 🔴 only Raft is really supported (not BFT, ) 🔴 </li>
            <li> 🔴 depending on the network configuration can get very centralized 🔴 </li>
            <li> 🔴 lots of components of the network, feels complicated 🔴 </li>
            <li> 🔴 verbose cli commands requiring lots of arguments (supposedly improved by v2.4 Fabric Gateway, yet the documentation is older than that version) 🔴 </li>
            <li> 🔴 unfortunately does not include an SDK in any of the civilized languages yet - go could work 🔴 </li>
            <li> 🔴 more code / work required compared to deploying smart contracts to an existing chain 🔴 </li>
            <li> 🔴 the least user-friendly (yet powerfull) CLI I have ever seen; looks like it lacks an another abstraction layer over it's internals; some examples use node to provide that abstraction layer - so it seems like that is also meant to be user-configurable 🔴 </li>
        </ul>
    </li>
        <li>Polkadot and Monero:
        <ul>
            <li> 🔴 Don't support smart contracts by design; they are possible only by using parachains connecting to the main chain - which probably means they are less refined/harder to work with 🔴 </li>
        </ul>
    </li>

</ul>

Ethereum is not considered because of the upcoming merge (currently scheduled in semptember and it may or may not break the application), as well as (currently?) ridiculous fees, transaction speed and scalability
