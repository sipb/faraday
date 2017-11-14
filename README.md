# Faraday

This is a system, based on wireguard, for connecting a variable number of
endpoints (such as containers) together over a secure VPN.

**NOTE: THIS SYSTEM IS INCOMPLETE; THE FEATURES LISTED HERE ARE PLANS, NOT A
CURRENT IMPLEMENTATION STATUS.**

This extends wireguard with tools to automatically update its configuration,
based on the addition and removal of endpoints in the cluster.

It also includes distributing the SSH keys used for wireguard, authenticating
them based on existing key management infrastructure.

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

DESIGN QUESTION: should the central server store the current cluster state:

  - In memory
  - On disk
  - In etcd

### Components

The different components of the Faraday system are as follows:

  * farametad: runs on a central server
  * faradayd: runs on each node, managing wireguard

Working here...

## License

The files in this particular project are licensed under the MIT license.

## Guidelines for Go code

Every commit with Go code should be formatted with gofmt if possible. If not,
sequences of commits should be formatted at the end.

# Contact

Project lead: cela. Contact over zephyr (-c hyades) or email @mit.edu.
