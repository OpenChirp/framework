Main Interface: [![Godoc](https://godoc.org/github.com/OpenChirp/framework?status.png)](https://godoc.org/github.com/OpenChirp/framework)

REST Interface: [![Godoc](https://godoc.org/github.com/OpenChirp/framework/rest?status.png)](https://godoc.org/github.com/OpenChirp/framework/rest)

PubSub Interface: [![Godoc](https://godoc.org/github.com/OpenChirp/framework/pubsub?status.png)](https://godoc.org/github.com/OpenChirp/framework/pubsub)

# Description
This is the Golang [User](user.go), [Device](device.go), and [Service](service.go) client library for the OpenChirp framework.

# Structure

## Top level functions
* User client interfaces are created using `framework.StartUserClient()`
* Device client interfaces are created using `framework.StartDeviceClient()`
* Service client interfaces are created using `framework.StartServiceClient()`

The [Client](client.go) class serves as the parent class of all the above client interfaces and should not be directly used.
The purpose of the clients are to provide a single uniform interface for all OpenChirp functionality. The client libraries combine the OpenChirp [REST](rest) and [PubSub](pubsub) protocols into a single abstraction.

## REST
The pure http rest interface is exposed as the Golang [rest](rest) package.

## PubSub
The pure pubsub(MQTT) interface is exposed as the Golang [pubsub](pubsub) package.
