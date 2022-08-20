# How to run documentation examples

Check out this articles for a quickstart with Hyperledger Sawtooth:

<ul>
    <li> Hyperledger Developer's Guide (with installation and network creation) : https://sawtooth.hyperledger.org/docs/1.2/app_developers_guide/</li>
    <li> Transactions and Batches : https://sawtooth.hyperledger.org/docs/1.2/architecture/transactions_and_batches.html </li>
</ul> 


Running the sawtooth demo described in Using Docker for a Single Sawtooth Node section of the developer's guide with docker: (devmode consensus algorithm)

```console
$ curl https://raw.githubusercontent.com/hyperledger/sawtooth-core/1-2/docker/compose/sawtooth-default.yaml > sawtooth-default.yaml
$ docker compose -f sawtooth-default.yaml up
```


Note that this demo also contains the tic-tac-toe game on the blockchain described here: <https://sawtooth.hyperledger.org/docs/1.2/app_developers_guide/intro_xo_transaction_family.html>

<br>
<br>

## Running the Sawtooth Network described in `Setting Up a Sawtooth Node for Testing` section of the documentation:

### With PoET consensus algorithm:

```console
$ curl https://raw.githubusercontent.com/hyperledger/sawtooth-core/1-2/docker/compose/sawtooth-default-poet.yaml > sawtooth-default-poet.yaml
$ perl -pe 's/\\\n//' sawtooth-default-poet.yaml > tmp.yaml && mv tmp.yaml sawtooth-default-poet.yaml
$ docker compose -f sawtooth-default-poet.yaml up
```

### With PBFT consensus algorithm:

```console
$ curl https://raw.githubusercontent.com/hyperledger/sawtooth-core/1-2/docker/compose/sawtooth-default-pbft.yaml > sawtooth-default-pbft.yaml
$ perl -pe 's/\\\n//' sawtooth-default-pbft.yaml > tmp.yaml && mv tmp.yaml sawtooth-default-pbft.yaml
$ docker compose -f sawtooth-default-pbft.yaml up
```

<!-- Or with kubernetes:

```console
$ curl https://raw.githubusercontent.com/hyperledger/sawtooth-core/1-2/docker/kubernetes/sawtooth-kubernetes-default-poet.yaml > sawtooth-kubernetes-default-poet.yaml
$ docker compose -f sawtooth-default-poet.yaml up
``` -->

Note: Documentation might be a bit outdated: docker-compose became docker compose. As a result newlines in the yaml files are not escaped properly (at least for me). I added a simple command to remove the escaped new lines from the files. This makes them unreadable, but at least they work. The first demo appears to be working properly even without it though.
