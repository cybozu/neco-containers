s3gw
====

s3gw is a minimal gateway implementation for Ceph RGW API.

Environment variables
---------------------

- `BUCKET_HOST`: object bucket API endpoint host
- `BUCKET_PORT`: object bucket API endpoint port
- `BUCKET_NAME`: object bucket name
- `BUCKET_REGION`: object bucket region
- `AWS_ACCESS_KEY_ID`: access key ID
- `AWS_SECRET_ACCESS_KEY`: access key secret

`BUCKET_HOST`,  `BUCKET_NAME`, `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` are required. `BUCKET_PORT` and `BUCKET_REGION` are optional.

Options
-------

- `--listen`: addr:port to listen to (passed to http.Server.Addr)
- `--use-path-style`: use path-style bucket name
- `--hosts-allow`: subnets allowed to access to this gw, separated by comma
- `--hosts-deny`: subnets denied to access to this gw, separated by comma
  - note: `--hosts-allow` and `--hosts-deny` control access to `/bucket/` only. `/health` and `/metrics` accept access from everywhere.

Supported API
-------------

- `/health`: health endpoint. note: it does not represent the upstream S3 API health.
- `/metrics`: metrics endpoint.
- `/bucket/`: bucket objects listing (see below)
- `/bucket/<object-key>`: bucket objects operation. you can `GET`/`PUT`/`DELETE` objects.

## bucket objects listing

`/bucket/` returns a JSON like this:

```
{
    "objects": [
        {
            "key": "some_object",
            "size": 123
        },
        {
            "key": "another_object",
            "size": 789
        },
    ]
}
```

If there are no objects, `objects` is empty array `[]`, not `null`.

Restrictions
------------

This software is originally designed to proxy requests to Ceph RGW. It may not work with real Amazon S3.
