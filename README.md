# Faraday

This is a system, based on wireguard, for connecting a variable number of
endpoints (such as containers) together over a secure VPN.

**NOTE: THIS SYSTEM IS INCOMPLETE; THE FEATURES LISTED HERE ARE PLANS, NOT A
CURRENT IMPLEMENTATION STATUS.**

This extends wireguard with tools to automatically update its configuration,
based on the addition and removal of endpoints in the cluster.

It also includes distributing the keys used for wireguard, authenticating them
based on existing key management infrastructure.

An integration with rkt is provided, so that containers can easily be added as
endpoints with no access to the network except through the wireguard interface.

## Architecture

Because Faraday is designed as an overlay network for a storage cluster, it
needs to have high performance and be resistant to failures of any individual
node.

Wireguard provides two obvious topologies: set up a central VPN server, to
which each peer connects, or configure each peer to speak with each other peer
directly.

The problem with the former design is that this requires that all traffic flow
through a single server, which would be prohibitively expensive when hundreds
of gigabits of data are flowing through the cluster, and would mean that any
failure of this central node would bring down the entire disk cluster.

The problem with the latter design is that it requires knowing the entire
network topology beforehand, making it difficult to add and remove servers from
the cluster.

Faraday is based on using wireguard in the latter topology, with a central
server used to coordinate updates to the cluster topology. This way, loss of
the central server will only prevent updates to the network topology, not cause
a networking failure. Daemons on the individual nodes relay information to and
from the central server to track and update the state of the cluster.

The central server is only used for tracking *additions* to the cluster, and
only tracks removals from the cluster for the sake of having up-to-date lists
to give to new cluster members. Removals from the cluster are handled on the
peer level: once a peer goes offline for long enough, the individual node will
ask the central server whether it is still present, and if not, remove it.
(This extra check makes sure that the cluster doesn't get into a state where
two nodes think the other is gone, but the central node thinks they're both
there, so the never realize the other one is back.)

Because the central server isn't responsible for node removal, it doesn't need
to store the state of the cluster in any non-volatile sense -- when it
restarts, it slowly reconstructs the cluster state from scratch as each node
regularly pings it. This means that it might take a bit for each new node in
the cluster to discover all of the other peers, when the central server has
just been restarted.

### Components

The different components of the Faraday system are as follows:

  * farad: runs on the central server
  * faradayd: runs on each node, managing wireguard
  * rkt integration: wraps faradayd to be used as the main network interface
    for an rkt container.

The nodes connect to the central server over TLS, and talk to each other over
TLS as well. Connections will be kept open where possible, and only
re-established when necessary.

### Responsibilities

For farad:

  * Take as configuration a TLS CA for verifying faradayd instances.
  * Take as configuration a TLS certificate for authenticate itself as a
    server.
  * Track the current state of the cluster in memory.
  * Accept connections from authenticated nodes, whether or not in the cluster.
  * Keep connections open over time.
  * Allow nodes to request to join the cluster, and record their membership.
  * Tell nodes when a new node has joined the cluster.
  * Allow nodes to check the current state of a particular other node in the
    cluster.
  * Allow nodes to notify farad that they are still awake, which should be
    optional if they have otherwise interacted with farad.
  * Detect when nodes have not interacted with farad recently, and remove them
    from the cluster state.
  * Propagate public keys of faradayd nodes between them.

For faradayd:

  * Take as configuration a TLS certificate for authenticating itself.
  * Take as configuration a TLS CA for verifying other faradayd instances.
  * Track the current state of the cluster in memory.
  * Open authenticated connections to farad.
  * Request to join the cluster.
  * Keep connections open over time.
  * Regularly remind farad that faradayd is still running.
  * Open authenticated connections with other instances of faradayd, and
    receive the corresponding connections from other instances.
  * Regularly ping each other instance of faradayd, and detect when it stops
    communicating.
  * Request verification from farad before removing a disconnected faradayd
    peer.
  * Create and maintain a wireguard network interface.
  * Continually reconfigure the wireguard network interface to match the
    current tracked cluster state.
  * Generate a wireguard private key if one does not already exist.
  * Tell farad about the public key corresponding to the wireguard private key.
  * Receive public keys for other nodes from farad and pass them to wireguard.

For rkt integration:

  * TODO

## License

The files in this particular project are licensed under the MIT license.

## Guidelines for Go code

Every commit with Go code should be formatted with gofmt if possible. If not,
sequences of commits should be formatted at the end.

# Contact

Project lead: cela. Contact over zephyr (-c hyades) or email @mit.edu.
