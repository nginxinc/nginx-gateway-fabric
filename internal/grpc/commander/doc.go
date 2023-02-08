/*
Package commander holds all the objects and methods for interacting with agents through the gRPC Commander Service.

This package includes:
- Commander: object that implements the Commander interface.
- Connector: interface for connector between CommandChannelServer and agent identifying information
- ConnectorManager: interface for managing connectors
- Connection: object that encapsulates a connection to an agent.
- BidirectionalChannel: object that encapsulates the bidirectional streaming channel: CommandChannelServer.
*/

package commander
