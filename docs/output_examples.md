# Output examples
This file provides examples of the expected XML output for various supported operations.

## List devices
```
<devices xmlns="urn:bbf:yang:bbf-device-aggregation">
  <device>
    <name>3ee22e94-da3b-4048-9137-fdd12fde138a</name>
    <type xmlns:bbf-dvct="urn:bbf:yang:bbf-device-types">bbf-dvct:olt</type>
    <data>
      <hardware xmlns="urn:ietf:params:xml:ns:yang:ietf-hardware">
        <component>
          <name>3ee22e94-da3b-4048-9137-fdd12fde138a</name>
          <hardware-rev/>
          <firmware-rev/>
          <serial-num>BBSIM_OLT_10</serial-num>
          <mfg-name>BBSim</mfg-name>
          <model-name>asfvolt16</model-name>
          <state>
            <admin-state>unlocked</admin-state>
            <oper-state>enabled</oper-state>
          </state>
        </component>
      </hardware>
    </data>
  </device>
  <device>
    <name>d0eee966-31dd-4d7b-af83-50681761f766</name>
    <type xmlns:bbf-dvct="urn:bbf:yang:bbf-device-types">bbf-dvct:onu</type>
    <data>
      <hardware xmlns="urn:ietf:params:xml:ns:yang:ietf-hardware">
        <component>
          <name>d0eee966-31dd-4d7b-af83-50681761f766</name>
          <parent>3ee22e94-da3b-4048-9137-fdd12fde138a</parent>
          <parent-rel-pos>536870912</parent-rel-pos>
          <hardware-rev/>
          <firmware-rev/>
          <serial-num>BBSM000a0001</serial-num>
          <mfg-name>BBSM</mfg-name>
          <model-name>v0.0.1</model-name>
          <state>
            <admin-state>unlocked</admin-state>
            <oper-state>enabled</oper-state>
          </state>
        </component>
      </hardware>
      <interfaces xmlns="urn:ietf:params:xml:ns:yang:ietf-interfaces">
        <interface>
          <name>BBSM000a0001-1</name>
          <type xmlns:bbf-xponift="urn:bbf:yang:bbf-xpon-if-type">bbf-xponift:onu-v-vrefpoint</type>
          <oper-status>up</oper-status>
        </interface>
        <interface>
          <name>BBSM000a0001-2</name>
          <type xmlns:bbf-xponift="urn:bbf:yang:bbf-xpon-if-type">bbf-xponift:onu-v-vrefpoint</type>
          <oper-status>unknown</oper-status>
        </interface>
        <interface>
          <name>BBSM000a0001-3</name>
          <type xmlns:bbf-xponift="urn:bbf:yang:bbf-xpon-if-type">bbf-xponift:onu-v-vrefpoint</type>
          <oper-status>unknown</oper-status>
        </interface>
        <interface>
          <name>BBSM000a0001-4</name>
          <type xmlns:bbf-xponift="urn:bbf:yang:bbf-xpon-if-type">bbf-xponift:onu-v-vrefpoint</type>
          <oper-status>unknown</oper-status>
        </interface>
      </interfaces>
    </data>
  </device>
</devices>
```

## Get provisioned services
```
<vlan-translation-profiles xmlns="urn:bbf:yang:bbf-l2-access-attributes">
  <vlan-translation-profile>
    <name>BBSM000a0001-1-hsia</name>
    <ingress-rewrite>
      <push-outer-tag>
        <vlan-id>900</vlan-id>
      </push-outer-tag>
      <push-second-tag>
        <vlan-id>900</vlan-id>
      </push-second-tag>
    </ingress-rewrite>
  </vlan-translation-profile>
</vlan-translation-profiles>
<line-bandwidth-profiles xmlns="urn:bbf:yang:bbf-nt-line-profile">
  <line-bandwidth-profile>
    <name>User_Bandwidth2</name>
    <fixed-bandwidth>100000</fixed-bandwidth>
    <assured-bandwidth>100000</assured-bandwidth>
    <maximum-bandwidth>100000</maximum-bandwidth>
  </line-bandwidth-profile>
  <line-bandwidth-profile>
    <name>User_Bandwidth1</name>
    <fixed-bandwidth>30000</fixed-bandwidth>
    <assured-bandwidth>100000</assured-bandwidth>
    <maximum-bandwidth>100000</maximum-bandwidth>
  </line-bandwidth-profile>
</line-bandwidth-profiles>
<service-profiles xmlns="urn:bbf:yang:bbf-nt-service-profile">
  <service-profile>
    <name>BBSM000a0001-1-hsia</name>
    <ports>
      <port>
        <name>BBSM000a0001-1</name>
        <port-vlans>
          <port-vlan>
            <name>BBSM000a0001-1-hsia</name>
          </port-vlan>
        </port-vlans>
        <technology-profile-id xmlns="urn:bbf:yang:bbf-nt-service-profile-voltha">64</technology-profile-id>
        <upstream-subscriber-bp-name xmlns="urn:bbf:yang:bbf-nt-service-profile-voltha">User_Bandwidth1</upstream-subscriber-bp-name>
        <downstream-subscriber-bp-name xmlns="urn:bbf:yang:bbf-nt-service-profile-voltha">User_Bandwidth2</downstream-subscriber-bp-name>
        <mac-learning-enabled xmlns="urn:bbf:yang:bbf-nt-service-profile-voltha">false</mac-learning-enabled>
        <dhcp-required xmlns="urn:bbf:yang:bbf-nt-service-profile-voltha">true</dhcp-required>
        <igmp-required xmlns="urn:bbf:yang:bbf-nt-service-profile-voltha">false</igmp-required>
        <pppoe-required xmlns="urn:bbf:yang:bbf-nt-service-profile-voltha">false</pppoe-required>
      </port>
    </ports>
  </service-profile>
</service-profiles>
```

## ONU Activated notification
```
<onu-state-change xmlns="urn:bbf:yang:bbf-xpon-onu-states">
  <detected-serial-number>BBSM000a0001</detected-serial-number>
  <channel-termination-ref>BBSIM_OLT_10-pon-0</channel-termination-ref>
  <onu-state-last-change>2022-07-13T13:27:35+00:00</onu-state-last-change>
  <onu-state xmlns:bbf-xpon-onu-types="urn:bbf:yang:bbf-xpon-more-types">bbf-xpon-onu-types:onu-present</onu-state>
  <detected-registration-id>3cf9f89c-457f-490a-9c1c-fec725d7a555</detected-registration-id>
</onu-state-change>
```