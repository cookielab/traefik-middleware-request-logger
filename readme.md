# Request and response logger

Add plugin into traefik static configuration

```yml
experimental:
  plugins:
    traefik-middleware-request-logger:
      moduleName: "github.com/cookielab/traefik-middleware-request-logger"
      version: "v0.0.9"
```

Add plugin into traefik via dynamic configuration (kubernetes)

```yml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
    name: my-traefik-middleware-request-logger
    namespace: my-namespace
spec:
    plugin:
        traefik-middleware-request-logger:
            ContentTypes:
                - application/json
            Limits:
                MaxBodySize: "1048576"
            RequestIDHeaderName: X-Request-ID
            SkipHeaders:
                - Authorization
            StatusCodes:
                - 200
```


Configuration:
When logging to stdout or stderr only the LogTarget value can be `stdout` or `stderr`.
When logging to url the LogTarget value must be `url` and the LogTargetUrl must be a valid url.


Example `stdout`:

```yml
---
ContentTypes: # log only these content types
  - application/json
Limits:
  MaxBodySize: 1048576 # max size of request/response body
RequestIDHeaderName: X-Request-ID # save uniq request id into this header
StatusCodes: # log only these status codes
  - 200
SkipHeaders:
  - Authorization
LogTarget: stdout # or "stderr"
```


Example `url`:
Note: `LogTarget` value must be `url` in and the `LogTargetUrl` must be a valid url.

```yml
---
ContentTypes: # log only these content types
  - application/json
Limits:
  MaxBodySize: 1048576 # max size of request/response body
RequestIDHeaderName: X-Request-ID # save uniq request id into this header
StatusCodes: # log only these status codes
  - 200
SkipHeaders:
  - Authorization
LogTarget: url
LogTargetUrl: https://consumer.logs.example.com/input
```
Conditions use "AND" (all conditions must be true). When request or response size exeed limit, the info string is present.
