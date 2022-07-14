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