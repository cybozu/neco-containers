bmc-log-collector
============================

`bmc-log-collector` collects hardware logs from Baseboard Management Controller (BMC) and output own stdout.

The following products are assumed as BMC.
- DELL integrated Dell Remote Access Controller (iDRAC) 

This program reads the "machineslist.json" and retrieves the System Event Log (SEL) and the Lifecycle (LC) log from each BMC. "bmc-log-collector" adds the serial, the node IP, and the log type (`SEL` or `LCLog`) to own STD output.

The LC log endpoint of iDRAC returns only the latest 50 entries (newest first), so the collector pages backward with the `$skip` query parameter until it reaches the entry read in the previous cycle. On the first collection for a machine, and after the LC log was cleared in iDRAC, only the latest page is emitted so that the whole history is not backfilled at once.

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
      --lclog-max-pages int          Maximum pages of the lifecycle log to read per scraping cycle (default 40)
      --machine-list-json string     Target machines list of log scraping (default "/config/machineslist.json")
      --pointer-dir-path string      Data directory of pointer management (default "/data/pointers")
      --scraping-interval-time int   Timer(sec) of scraping interval time (default 300)
      --user-id string               User ID of bmc-user-json JSON file (default "support")
```
