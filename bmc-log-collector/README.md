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
      --pointer-configmap string      Name of the ConfigMap in the namespace of the pod to store pointers (for --pointer-store=configmap) (default "bmc-log-collector-pointers")
      --pointer-dir-path string       Data directory of pointer management (for --pointer-store=file) (default "/data/pointers")
      --pointer-store string          Where to store pointers: "file" or "configmap" (default "file")
      --scraping-interval-timer int   Timer(sec) of scraping interval time (default 300)
      --user-id string                User ID of bmc-user-json JSON file (default "support")
```

## Pointer store

bmc-log-collector records the last read SEL entry of each BMC as a "pointer" so that it does not output the same entry twice.
The pointers can be stored in one of the following places, selected by `--pointer-store`.

- `file` (default): one file per serial under `--pointer-dir-path`. Mount a persistent volume on the directory.
- `configmap`: one key per serial in the ConfigMap named `--pointer-configmap` in the namespace of the pod.
  The pointers are kept in memory and written to the ConfigMap once per scraping interval with server-side apply.
  Run only one replica. The service account needs `get` and `patch` on the ConfigMap, and `create` if the ConfigMap does not exist yet.

Switching the store does not migrate the pointers, so all the SEL entries are output once again after switching.
