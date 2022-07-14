# BBF Adapter code structure

## `build` directory and Docker images

The `build` directory contains the files used for the creation of a voltha-northbound-bbf-adapter Docker image that can be used for deployment.

- `build/tools` contains the Dockerfile of the "builder" image. This image includes a golang toolchain and the .deb packages for libyang and sysrepo.\
The presence of these packages allows the correct compilation of the project, which would otherwise fail due to missing dependencies. The same image is used for linting and testing through the Makefile.
- `build/package` contains the Dockerfile of the actual BBF adapter image. It uses the builder image to compile the code of this repository. The generated binaries are then copied on the production image, where packages for libyang, sysrepo and netopeer2 are installed.\
During this step, the yang files in `build/yang-files` are installed with sysrepoctl, making them available at runtime.\
A user account that can be used to log in to netopeer2 is also created at this time. The credentials of this user can be changed by setting the `NETCONF_USER` and `NETCONF_PASSWORD` arguments.

When the adapter image is deployed with the `onf/voltha-northbound-bbf-adapter` helm chart defined in [voltha-helm-charts](https://github.com/opencord/voltha-helm-charts), netopeer2 is automatically started with the `netopeer2-server` command before running the adapter itself.

## `cmd` directory

The `cmd/bbf-adapter` directory contains the source for the main function of the adapter.\
`main.go` loads the configuration provided through CLI flags by the user, starts all the necessary components defined in the `internal` directory, keeps track of the adapter's readiness status and exposes it through the probe package of voltha-lib-go.

## `internal` directory

The `interal` directory contains the remaining logic of the adapter, split in multiple packages.

### `internal/clients`

This package contains the implementation of clients that the adapter can use to interact with the VOLTHA components that are necessary to provide its funcionality.
-  `nbi.go` provides a connection to VOLTHA's "northbound interface" gRPC service, which can be used to interact with the core the same way a user would using voltctl. This client can be used to list the devices, provision a new OLT and enable it.
- `olt_app.go` provides wrapper functions for the REST APIs exposed by the Olt ONOS application. This client can be used to provision a subscriber on a specific device and port.

### `internal/config`

This package contains the definition of CLI flags that can be used to configure the adapter's behavior, and their default values.

### `interal/core`

This package contains the implementation of the VolthaYangAdapter (`adapter.go`), an instance of which is exposed globally as core.AdapterInstance. This is necessary because the callback functions defined in `internal/sysrepo/syrepo.go` would have no other way to reference it when being called from sysrepo.\
VolthaYangAdapter provides methods to perform actions through its clients or request data from them. In the latter case, the functions defined in `translation.go` are used to translate information to a set of yang paths and values that will be used to update the NETCONF datastore.

### `internal/sysrepo`

Since sysrepo doesn't provide official bindings for golang, this project uses CGO to interact with sysrepo through its official C library: libsysrepo.

- `plugin.c` is a C file used to include dependencies (like libyang and sysrepo header files) and define some utilities that make it possible to perform all the necessary operations through CGO.
- `sysrepo.go` imports `plugin.c` and defines the SysrepoPlugin struct, which keeps track of the adapter's connection with sysrepo. At startup, a connection is created and the necessary callback functions are registered.
- `callbacks.go` contains the definitions of callback functions called by sysrepo for the subscribed operations
- `utils.go` contains functions to ease the interaction with Sysrepo and CGO

When a NETCONF operation is executed on one of the modules managed by the BBF adapter, the corresponding go function will be called. The latter will interact with the global instance of VolthaYangAdapter to fulfill the request and provide a result, eventually updating the content of the datastore.