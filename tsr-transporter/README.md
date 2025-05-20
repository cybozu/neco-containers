# TSR Transporter (tsr-transporter)

The TSR transporter receives TSR requests from Kintone apps, obtains TSRs from BMC, and registers them in Kintone apps.

## Usage 

```
This command act to get TSR from iDRAC and put in the record of Kintone:

Using the service tag of the server registered in the Kintone app as the key, 
find the IP address of the server's iDRAC/BMC (Baseboard Management Controller) from Sabakan, 
request a TSR (Technical Service Report) job, and register the obtained TSR in Kintone.

Usage:
  tsr-transporter [flags]
  tsr-transporter [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     print version of this command

Flags:
  -b, --bmc-user string   BMC user config
  -h, --help              help for tsr-transporter
  -k, --kintone string    Kintone App config
  -s, --sabakan string    Sabakan config

Use "tsr-transporter [command] --help" for more information about a command.
```

## Referenced files

### BMC user config (-b, --bmc-user string)
Set the support user ID and password for BMC.

```
{
    "root": {
        omit
    },
    "repair": {
        omit
    },
    "power": {
        omit
    },
    "support": {
      "password": {
        "raw": "set password"
      }
    }
}
```

### Kintone App config  (-k, --kintone string)
In the development and testing phase, set the address for the evaluation version of Kintone.
In the operation phase, set the parameters for the production version of the Kintone app.

```
{
    "domain": "must set domain of Kintone (string)",
    "app_id": must set Application ID (int),
    "space_id": must set space ID (int),
    "is_guest": true (bool),
    "proxy": "option (string)",
    "token": "must set token to access kintone app, that must have the read/write permissions",
    "working_dir": "must set working directory"
}
```

### Sabakan config
for development & test to use sabakan mock

```
{
    "service": "127.0.0.1:7180",
    "api_path": "/api/v1/machines"
}
```

for stage or production environment. must change namespace 
```
{
    "service": "sabakan.replace_your_namespace.svc",
    "api_path": "/api/v1/machines"
}
```
