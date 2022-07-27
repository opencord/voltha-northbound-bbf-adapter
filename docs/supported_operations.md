# Supported operations
This page provides a list of operations that are currently supported by the BBF Adapter, with references on how to perform them and documentations on their behavior.

## "Get" operations

### Get devices data
Information on the devices managed by VOLTHA can be retrieved with a `get-data` NETCONF operation on the `operational` datastore.\
The following XPath can be used to filter this information: `/bbf-device-aggregation:*`

An example of the exposed information can be found in [output_examples.md](output_examples.md) 

>The following information is not currently available but planned for future updates:
>- ONU software images
>- OLT endpoint information

### Get services data
Information on the provisioned services can be retrieved with a `get-data` NETCONF operation on the `operational` datastore.\
The following XPath can be used to filter this information: `/bbf-nt-service-profile:*|/bbf-l2-access-attributes:*|/bbf-nt-line-profile:*`

An example of the exposed information can be found in [output_examples.md](output_examples.md) 

>The translation of bandwidth profiles to YANG data is currently under discussion and will be provided in a future update

## "Set" operations

### Activate a service
A service can be activated on a specific UNI with the creation of nodes through an `edit-config` operation on the `running` datastore.\
The necessary information for the activation of a service are the UNI port name, C-Tag, S-Tag and Technology Profile ID.\
Configuration for both `vlan-translation-profiles` and `service-profiles` has to be created with a single request, since the operation will be translated to a single API call to ONOS. Failing to provide both will result in an error.

>The complete configuration for the service, matching the provided C-Tag, S-Tag and Technology Profile must be available to ONOS through SAIDS.

An example of the configuration to activate a service can be found in [provision_service.xml](../examples/provision_service.xml)

### Deactivate a service
A service can be deactivated on a specific UNI with the deletion of nodes through an `edit-config` operation on the `running` datastore.\
The necessary information for the deactivation of a service is the name used for its creation.\

An example of the configuration to deactivate a service can be found in [remove_service.xml](../examples/remove_service.xml)

## Notifications

### ONU Activated notification

A notification for the `ONU_ACTIVATED` event can be received by subscribing to the `bbf-xpon-onu-states` stream.
After this notification is received, services can be provisioned on the ONU it refers to.

An example of this notification can be found in [output_examples.md](output_examples.md) 

>The use of the bbf-xpon-onu-states yang module is temporary, and will be substituted after the definition of a VOLTHA specific yang notification