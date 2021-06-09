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

`BUCKET_HOST`, `BUCKET_NAME`, `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` are required. `BUCKET_PORT` and `BUCKET_REGION` are optional.

Options
-------

- `--listen`: addr:port to listen to (passed to http.Server.Addr)
- `--use-path-style`: use path-style bucket name
- `--hosts-allow`: subnets allowed to access to this gw, separated by comma
- `--hosts-deny`: subnets denied to access to this gw, separated by comma
  - note: `--hosts-allow` and `--hosts-deny` control access to `/bucket/` only. `/health` and `/metrics` accept access from everywhere.

Supported API
-------------

Unless otherwise stated, only `GET` method is supported. Other methods may return `405 Method not Allowed` or work as of `GET`.

### `/health`: health endpoint

It returns the health status of this GW.

note: it does not return the health status of the upstream S3 API.

### `/metrics`: metrics endpoint

In addition to standard promhttp metrics which begin with `go_` and `promhttp_`,  it returns the following metrics:

- `s3gw_request_count_total` counter
- `s3gw_request_duration_seconds` histogram

both of those metrics have the following labels:

- `code`: HTTP status code
- `method`: HTTP request method
- `handler`: request handler type. `"list"` for bucket objects listing requests or `"object"` for bucket objects operation.

### `/bucket/`: bucket objects listing

It returns the list of objects stored in the bucket as the following JSON format:

```json
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

### `/bucket/<object-key>`: bucket objects operation

You can `GET`/`PUT`/`DELETE` objects.

As with S3 API, `<object-key>` is not a path-style string. i.e. `/bucket/foo` and `/bucket//foo` represent different objects.

Restrictions
------------

This software is originally designed to proxy requests to Ceph RGW. It may not work with real Amazon S3.
