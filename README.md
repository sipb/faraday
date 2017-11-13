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

## License

The files in this particular project are licensed under the MIT license.

## Guidelines for Go code

Every commit with Go code should be formatted with gofmt if possible. If not,
sequences of commits should be formatted at the end.

If the code is part of a security-critical component, there should be
reasonably complete unit tests before merging.

# Contact

Project lead: cela. Contact over zephyr (-c hyades) or email @mit.edu.
