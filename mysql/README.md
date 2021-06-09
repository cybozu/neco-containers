# MySQL container

This directory provides a Dockerfile to build a MySQL container for [MOCO](https://github.com/cybozu-go/moco).
This also provides `moco-init` command to initialize MySQL data for MOCO.

## MOCO MySQL container

### Usage

This container image is assumed to be used by [MOCO](https://github.com/cybozu-go/moco).

### Docker images

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/mysql).

## moco-init command

### description

`moco-init` initializes the data directory with `mysqld --initialize-insecure`.
This mysqld command creates an initial root user with an empty password.

`moco-init` then makes a supplementary file for `my.cnf` with the following information.

* [`server_id`](https://dev.mysql.com/doc/refman/8.0/en/replication-options.html#sysvar_server_id)
* [`admin_address`](https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_admin_address)

`moco-init` is to be called as an init container at every startup of the Pod so that it prepares the data directory before starting `mysqld`.

The data directory is created under the base directory with the name `data`.

### synopsis

```
moco-init [options...] <server_id_base>
```

`server_id_base` is an integer to calculate `server_id` value.

```
<server_id> = <server_id_base> + <index>
  where Pod name is moco-<cluster_name>-<index> 
```

### options

| Name       | Default             | Description                                        |
| ---------- | ------------------- | -------------------------------------------------- |
| `base-dir` | `/usr/local/mysql`  | The directory where MySQL is installed.            |
| `data-dir` | `/var/mysql`        | The directory for MySQL data.                      |
| `conf-dir` | `/etc/mysql-conf.d` | The directory where configuration file is created. |

### environment variables

| Name       | Required | Description                 |
| ---------- | -------- | --------------------------- |
| `POD_NAME` | yes      | The name of the MySQLd Pod. |
