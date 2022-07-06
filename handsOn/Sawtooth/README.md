

Requirements:
<ul>
    <li> Tested on Ubuntu focal 20.04</li>
    <li> Rust installed                         (I've decided to use Rust API with Sawtooth)</li>
    <li> Docker engine and compose installed    (or setting up nodes for testing by hand) </li>
</ul>
 

References:
<ul>
    <li> Hyperledger Developer's Guide (with installation and network creation) : https://sawtooth.hyperledger.org/docs/1.2/app_developers_guide/</li>
</ul> 

Running the sawtooth demo described in Using Docker for a Single Sawtooth Node section of the developer's guide:
docker compose -f sawtooth-default.yaml up