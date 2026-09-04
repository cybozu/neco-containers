bmc-log-collector
============================

`bmc-log-collector` collects hardware logs from Baseboard Management Controller (BMC) and outputs them to its own stdout.

The following products are assumed as BMC.
- DELL integrated Dell Remote Access Controller (iDRAC) 

This program reads the "machineslist.json" and retrieves the System Event Log (SEL) and the Lifecycle (LC) log from each BMC. "bmc-log-collector" adds the serial, the node IP, and the log type (`SEL` or `LCLog`) to each entry and writes it to stdout.

The LC log endpoint of iDRAC returns only the latest 50 entries (newest first), so the collector follows `Members@odata.nextLink` backward until it reaches the entry read in the previous cycle, up to `--lclog-max-pages` pages per cycle. The first collection for a machine and the collection after the LC log was cleared in iDRAC go through the same loop; the page limit bounds the backfill so that the whole history is not ingested at once.

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
      --lclog-max-pages int          Maximum pages of the lifecycle log to read per scraping cycle (default 3)
      --machine-list-json string     Target machines list of log scraping (default "/config/machineslist.json")
      --pointer-dir-path string      Data directory of pointer management (default "/data/pointers")
      --scraping-interval-time int   Timer(sec) of scraping interval time (default 300)
      --user-id string               User ID of bmc-user-json JSON file (default "support")
```
