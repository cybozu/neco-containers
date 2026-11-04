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
      --scraping-interval-time int    Timer(sec) of scraping interval time (default 300)
      --startup-lookback duration     On start, output only SEL entries created within this duration before the first request to each BMC (default 24h0m0s)
      --user-id string                User ID of bmc-user-json JSON file (default "support")
```

## Duplicates and losses of logs

bmc-log-collector keeps the last read SEL entry of each BMC (the "pointer") only in memory, so it does not need any storage.
On start, and when a new machine appears in the machines list, it outputs only the entries created within `--startup-lookback` before the request to the BMC.
After that, it outputs only the entries newer than the pointer.

As a result, around a restart:

- Entries created within `--startup-lookback` before the restart may be output twice.
- Entries are lost if bmc-log-collector stays down for longer than `--startup-lookback`, or if the clock of the BMC is wrong.

Set `--startup-lookback` long enough to cover the downtime of a rolling update plus the scraping interval,
so that no entries are lost normally. Losses are acceptable only when bmc-log-collector stays down for a long time.

To handle the duplicates, use the `Created` field of each entry as the timestamp of the log in the log collector,
and deduplicate the entries with the same timestamp and content in the log viewer.
