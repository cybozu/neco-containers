bmc-log-collector
============================

`bmc-log-collector` collects hardware logs from Baseboard Management Controller (BMC).

The following products are assumed as BMC.
- DELL integrated Dell Remote Access Controller (iDRAC) 

This program reads the "server-list" in JSON format and retrieves the System Event Log from each BMC. "bmc-log-collector" adds the serial and the node name to own output. the serial and the node are from the "server-list" 


## server-list JSON file

server-list includes serial, BMC IP address, Server IP address in below JSON format.

```
[
    {
        serial: "ABC123",
        bmc_ip: "192.168.1.1"
        ipv4:   "192.168.10.1"
    },
    {
        // Next server 
    },
    // Repeat
]
```

## Usage 

 *** under writing ***
