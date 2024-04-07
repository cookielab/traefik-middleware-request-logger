# Request/response logger

file provider example:

```yml
http:
  middlewares:
    my-plugin:
      plugin:
        traefik-middleware-request-logger:
          RequestIDHeaderName: X-Request-Id
          ContentTypes:
            - application/json
          StatusCodes:
            - 200
            - 500
          Limits:
            MaxBodySize: 16384
```

crd example:

```yml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: log-request
spec:
  plugin:
    traefik-middleware-request-logger:
      RequestIDHeaderName: X-Request-Id
      ContentTypes:
        - application/json
      StatusCodes:
        - 200
        - 500
      Limits:
        MaxBodySize: 16384
```

configMap via helm chart

```yml
apiVersion: v1
kind: ConfigMap
metadata:
  name: traefik-middleware-request-logger
data:
{{ (.Files.Glob "traefik-middleware-request-logger/*").AsConfig | indent 2 }}
```

traefik helm chart values example (local plugin mode):

```yaml
additionalArguments:
  - >-
    --experimental.localplugins.log-request.modulename=github.com/cookielab/traefik-middleware-request-logger
additionalVolumeMounts:
  - mountPath: /plugins-local/src/github.com/cookielab/traefik-middleware-request-logger
    name: plugins
deployment:
  additionalVolumes:
    - configMap:
        name: traefik-middleware-request-logger
        items:
          - key: dot.traefik.yml
            path: .traefik.yml
          - key: go.mod
            path: go.mod
          - key: logger.go
            path: logger.go
      name: plugins
```