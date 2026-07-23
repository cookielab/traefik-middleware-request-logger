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
RedactBodyFields: # body fields whose values are replaced with "[REDACTED]"
  - password
  - token
LogTarget: stdout # or "stderr"
```

Body redaction is **on by default**: when `RedactBodyFields` is not set, a built-in
list of sensitive field names is used (`password`, `passwd`, `pwd`, `secret`,
`client_secret`, `token`, `access_token`, `refresh_token`, `id_token`, `api_key`,
`authorization`, `authorizationCode`, and camelCase variants). Matching is
case-insensitive and applies to JSON bodies (recursively) as well as bodies logged
as plain strings (`"field": "..."` and `field=...` shapes). Set `RedactBodyFields`
to override the list with your own.


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
