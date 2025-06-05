gohttp
===

gohttp is the simplest way to serve static files with high performance.

# conf.yaml

gohttp expects a `conf.yaml` in the same folder, example:

```yaml
port: 80
domain: localhost
staticpath: www
cors: '*'
```

 where `www` is the folder you want to serve on the port
