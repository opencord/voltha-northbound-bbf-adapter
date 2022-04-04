# Expanding the adapter's functionality

## Resources
- [libsysrepo API documentation](https://netopeer.liberouter.org/doc/sysrepo/devel/html/modules.html)
- [libyang API documentation](https://netopeer.liberouter.org/doc/libyang/devel/html/modules.html)

## Schema-mount support
The yang modules used to map VOLTHA APIs use the schema-mount yang extension to create mountpoints, in which other moudels can be included. Libyang is the library used by sysrepo and netopeer2 to operate on yang data, and for this reason it needs to support the extension to let the BBF adapter create all the needed leafs.

To make the extension work, additional data has to be provided to libyang (by both the adapter and netopeer2) through a callback function. More information on the data can be found [here](https://netopeer.liberouter.org/doc/libyang/devel/html/group__context.html#ga14853fe1a338c94d9e81be9566438243).

The set of mountpoints used by the adapter (which has to be modified in case a new yang file with a mountpoint is added) is defined with an XML file, which can be found in `build/yang-files/schema-mount.xml`. This file is copied into the Docker image of the adapter, and is automatically extended with the `ietf-yang-library` provided by sysrepo.
The path of `schema-mount.xml` is then passed to each process in [the startup script](https://github.com/opencord/voltha-helm-charts/blob/master/voltha-northbound-bbf-adapter/templates/configmap-startup.yaml).

### Schema-mount callback in the BBF adapter
The BBF adapter loads the content of `schema-mount.xml` (or its equivalent, whose path is provided with the `--schema_mount_path` flag) as libyang nodes during the startup of its sysrepo plugin.
The callback that will pass this data to libyang is defined in `internal/sysrepo/plugin.c`, and is also registered during the startup.

## How to expand the adapter's functionality by creating new callbacks

- If additional yang modules are needed, add them to `build/yang-files` to be installed to sysrepo during the creation of the adapter's Docker image
- If additional sysrepo configuration is needed for the moudule (i.e. enabling a feature) it can be added to `build/package/Dockerfile.bbf-adapter`
- A new callback function has to be created. CGO cannot directly express the `const` keyword used in the signature of the C callbacks of libsysrepo.\
For this reason, a Go function can be created and exposed in `internal/sysrepo/sysrepo.go`, while its corresponding wrapper with the right signature has to be created in `internal/sysrepo/plugin.go`.
- The new callback wrapper has to be registered in the StartNewPlugin function of `internal/sysrepo/sysrepo.go`, using the sysrepo API to subscribe to the desired type of operation and XPath.
- The Golang callback will interact with VOLTHA through the instance of VolthaYangAdapter exposed by the `core.AdapterInstance` variable.\
If the operation needed by the new callback is not implemented already, it can be created as a new method in `internal/core/adapter.go`, eventually expanding the capabilities of the clients in `internal/clients/nbi.go` and `internal/clients/olt_app.go`, or creating new translation functions in `internal/core/translation.go`.