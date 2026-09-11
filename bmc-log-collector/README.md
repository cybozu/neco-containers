bmc-log-collector
============================

`bmc-log-collector` collects hardware logs from Baseboard Management Controller (BMC) and outputs them to its own stdout.

The following products are assumed as BMC.
- DELL integrated Dell Remote Access Controller (iDRAC) 

This program reads the "machineslist.json" and retrieves the System Event Log (SEL) and the Lifecycle (LC) log from each BMC. "bmc-log-collector" adds the serial, the BMC IP, the node IP, and the log type (`SEL` or `LCLog`) to each entry and writes it to stdout.

The LC log endpoint of iDRAC returns only the latest 50 entries. The collector emits the entries of that page which are newer than the one read in the previous cycle; the entries that fell off the page since then are not collected (the LC log grows only a few entries per day in our fleet).

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
      --bmc-user-json string         User and password of BMC (default "/users/neco/bmc-user.json")
      --machine-list-json string     Target machines list of log scraping (default "/config/machineslist.json")
      --pointer-dir-path string      Data directory of pointer management (default "/data/pointers")
      --scraping-interval-time int   Timer(sec) of scraping interval time (default 300)
      --user-id string               User ID of bmc-user-json JSON file (default "support")
```
