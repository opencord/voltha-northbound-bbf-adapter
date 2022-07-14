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

Running the following instruction will connect to the adapter's netopeer2 instance as the default `voltha` user.

```
connect --login voltha
```

When presented with the server's fingerprint, confirm by entering `yes`, and then log in with password `onf`.

After a successful login, requests can be performed.

### Getting device data

```
get-data --datastore operational --filter-xpath /bbf-device-aggregation:*
```

### ONU activation notifications

Run the following commands inside the netopeer2-cli console.
```
ext-data /schema-mount.xml
subscribe --stream bbf-xpon-onu-states
```
A notification will be shown when a new ONU is activated.

## Stop the BBF adapter
```
helm delete -n voltha bbf
```
