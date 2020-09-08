Elasticsearch container
==================

Build Docker container image for [Elasticsearch][], which is a full text search engine

Usage
-----

### Run elasticsearch at local:

```console
$ docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" quay.io/cybozu/elasticsearch:7.9.1.1
```

[Elasticsearch]: http://github.com/elastic/elasticsearch/

Docker images
-------------

Docker images are available on [Quay.io](https://quay.io/repository/cybozu/elasticsearch)
