bmc-log-collector
============================

`bmc-log-collector` collects hardware logs from Baseboard Management Controller (BMC) and output own stdout.

The following products are assumed as BMC.
- DELL integrated Dell Remote Access Controller (iDRAC) 

This program reads the "machineslist.json" and retrieves the System Event Log from each BMC. "bmc-log-collector" adds the serial and the node IP to own STD output.

## Referenced file

#### User and password of BMC

```
{
  "USERID-TO-BE-REPLACE": {
    "password": {
      "raw": "PASSWORD-STRING-TO-BE-REPLACE"
    }
  },
  // Repeat
}
```

#### Target "machineslist.json" of log scraping

```
[
    {
        serial:    "ABC1234",     // Uniq serial ID of the server hardware
        bmc_ipv4:  "192.168.1.1"  // BMC IP address
        node_ipv4: "192.168.10.1" // Server IP address
    },
    // Repeat
]
```


## Usage 

bmc-log-collector command provides the usage in following.

```
$ bmc-log-collector --help

Usage of ./bmc-log-collector:
      --bmc-user-json string          User and password of BMC (default "/users/neco/bmc-user.json")
      --machine-list-json string      Target machines list of log scraping (default "/config/machineslist.json")
      --pointer-dir-path string       Data directory of pointer management (default "/data/pointers")
      --scraping-interval-timer int   Timer(sec) of scraping interval time (default 300)
      --user-id string                User ID of bmc-user-json JSON file (default "support")
```
