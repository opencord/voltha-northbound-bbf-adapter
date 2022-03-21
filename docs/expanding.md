# Expanding the adapter's functionality

## Resources
- [libsysrepo API documentation](https://netopeer.liberouter.org/doc/sysrepo/master/html/modules.html)
- [libyang API documentation](https://netopeer.liberouter.org/doc/libyang/master/html/modules.html)

## How to expand the adapter's functionality by creating new callbacks

- If additional yang modules are needed, add them to `build/yang-files` to be installed to sysrepo during the creation of the adapter's Docker image
- A new callback function has to be created. CGO cannot directly express the `const` keyword used in the signature of the C callbacks of libsysrepo.\
For this reason, a Go function can be created and exposed in `internal/sysrepo/sysrepo.go`, while its corresponding wrapper with the right signature has to be created in `internal/sysrepo/plugin.go`.
- The new callback wrapper has to be registered in the StartNewPlugin function of `internal/sysrepo/sysrepo.go`, using the sysrepo API to subscribe to the desired type of operation and XPath.
- The Golang callback will interact with VOLTHA through the instance of VolthaYangAdapter exposed by the `core.AdapterInstance` variable.\
If the operation needed by the new callback is not implemented already, it can be created as a new method in `internal/core/adapter.go`, eventually expanding the capabilities of the clients in `internal/clients/nbi.go` and `internal/clients/olt_app.go`, or creating new translation functions in `internal/core/translation.go`.