# Deployment of the BBF adapter

Similarly to other components of the VOLTHA stack, the BBF adapter is deployed with a helm chart.

## Prerequisites

BBF adapter requires a working deployment of the voltha-infra and voltha-stack helm charts for a successful installation.
Please refer to [docs.voltha.org](https://docs.voltha.org/master/voltha-helm-charts/README.html) to learn how to set them up.

## Install the voltha-northbound-bbf-adapter helm chart

The adapter's chart can be installed with the following command, assuming the voltha-infra and voltha-stack charts have been deployed with suggested names from [docs.voltha.org](https://docs.voltha.org/master/voltha-helm-charts/README.html).

```
helm upgrade --install --create-namespace -n voltha bbf \
            onf/voltha-northbound-bbf-adapter --devel \
            --set global.voltha_infra_name=voltha-infra \
            --set global.voltha_infra_namespace=infra \
            --set global.voltha_stack_name=voltha \
            --set global.voltha_stack_namespace=voltha \
            --set global.log_level=INFO \
            --set images.voltha_northbound_bbf_adapter.tag=master
```

If needed, the SSH port on which netopeer2 is listening can be exposed on localhost by running:

```
kubectl -n voltha port-forward svc/bbf-voltha-northbound-bbf-adapter-netopeer2 50830
```

The logs of the adapter can be followed in a separate terminal:

```
kubectl -n voltha logs --follow $(kubectl -n voltha get pods -l app=bbf-adapter -o name)
```

## Make NETCONF requests
After a succesful installation, a NETCONF client can be used to perform requests to the adapter's netopeer2 instance.\
For these examples, we will use an instance of netopeer2-cli  running inside the BBF Adapter's container.

```
kubectl -n voltha exec -it $(kubectl -n voltha get pods -l app=bbf-adapter -o name) -- netopeer2-cli
```

Running the following instructions will connect to the adapter's netopeer2 instance as the default `voltha` user.

```
searchpath /etc/sysrepo/yang
ext-data /schema-mount.xml
connect --login voltha
```

When presented with the server's fingerprint, confirm by entering `yes`, and then log in with password `onf`.

After a successful login, requests can be performed.

### Getting device data

Run the following commands inside the netopeer2-cli console.
```
get-data --datastore operational --filter-xpath /bbf-device-aggregation:*
```

### ONU activation notifications

Run the following commands inside the netopeer2-cli console.
```
subscribe --stream bbf-xpon-onu-states
```
A notification will be shown when a new ONU is activated.

### Provision and remove a service

Run the following command in a separate terminal, from the root of this repository, to copy the example XMLs into the adapter's container.
```
kubectl cp examples/ voltha/$(kubectl -n voltha get pods -l app=bbf-adapter -o name | awk -F "/" '{print $2}'):/
```
To provision the service with one of the example XMLs, run the following command inside the netopeer2-cli console.
```
edit-config --target running --config=/examples/provision_service.xml
```
The details of the provisioned services can be retrived with the following command.
```
get-data --datastore operational --filter-xpath /bbf-nt-service-profile:*|/bbf-l2-access-attributes:*|/bbf-nt-line-profile:*
```
Finally, the service can be removed.
```
edit-config --target running --config=/examples/remove_service.xml
```

## Stop the BBF adapter
```
helm delete -n voltha bbf
```
