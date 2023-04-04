/*
Package agent contains objects and methods for configuring agents.

The package includes:
- ConfigStore: a thread-safe store for latest agent nginx configuration.
- NginxConfigBuilder: builds agent nginx configuration from dataplane.Configuration.
- NginxConfig: an intermediate object that contains nginx configuration in a form that agent expects.
*/
package agent
