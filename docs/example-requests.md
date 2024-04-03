# Example requests

Example request for common API requests.

Request can either be performed against the local API server at `localhost:3000` or against the public release of the API server at [status.k8s.scs.community](https://status.k8s.scs.community/). When using the public release authorization via bearer token must be included. Please refer to the [status page deployment](https://github.com/SovereignCloudStack/status-page-deployment).
 Local instances of the API can use `STATUS_PAGE_SWAGGER_UI_ENABLED=true` to perform request with swagger at `localhost:3000/swagger`.

All examples include the public release with authorization header.

## List phases

```bash
$ curl -sL \
https://status.k8s.scs.community/phases

{"data":{"generation":1,"phases":["Scheduled","Investigation ongoing","Working on it","Potential fix deployed","Done"]}}
```

## Create phases

```bash
$ curl -sL \
-X POST \
-H 'Authorization: Bearer <your-id-token>' \
-H 'Content-Type: application/json' \
-d '{"phases": ["Phase 1", "Phase 2", "Phase 3"]}' \
https://status.k8s.scs.community/phases

{"generation":2}
```

## List impact types

```bash
$ curl -sL \
https://status.k8s.scs.community/impacttypes

{"data":[{"displayName":"Performance Degration","id":"63645189-ffbc-4e6c-b991-99e14ade3edc"},{"displayName":"Connectivity Problems","id":"9fcf0039-9e24-45b0-b76a-80e6246a803b"},{"displayName":"Unknown","id":"6cadd69a-5702-4824-b7a6-d5509c54b8cb"}]}
```

## Create impact type

```bash
$ curl -sL \
-X 'POST' \
-H 'Authorization: Bearer <your-id-token>' \
-H 'Content-Type: application/json' \
-d '{"displayName": "Test impact type"}' \
https://status.k8s.scs.community/impacttypes

{"id":"52d178d2-0fe9-4654-834b-42502283454d"}
```

## Get impact type

```bash
$ curl -sL \
https://status.k8s.scs.community/impacttypes/52d178d2-0fe9-4654-834b-42502283454d

{"data":{"displayName":"Test impact type","id":"52d178d2-0fe9-4654-834b-42502283454d"}}
```

## List severities

```bash
$ curl -sL \
https://status.k8s.scs.community/severities

{"data":[{"displayName":"operational","value":25},{"displayName":"maintenance","value":50},{"displayName":"limited","value":75},{"displayName":"broken","value":100}]}
```

## Create severity

```bash
curl -sL \
-X POST \
-H 'Authorization: Bearer <your-id-token>' \
-H 'Content-Type: application/json' \
-d '{"displayName": "test","value": 60}'
https://status.k8s.scs.community/severities
```

## Get severity

```bash
$ curl -sL \
https://status.k8s.scs.community/severities/test

{"data":{"displayName":"test","value":60}}
```

## List components

```bash
$ curl -sL \
https://status.k8s.scs.community/components

{"data":[{"activelyAffectedBy":[],"displayName":"Storage","id":"871584dc-e425-4155-8fde-47ca588689f3","labels":{}},{"activelyAffectedBy":[],"displayName":"Network","id":"414d764f-0c94-4c4d-90e6-c97cce00cce3","labels":{}},...]}
```

## Create component

```bash
curl -sL \
-X POST \
-H 'Authorization: Bearer <your-id-token>' \
-H 'Content-Type: application/json' \
-d '{"displayName":"Test-Component"}'\
https://status.k8s.scs.community/components

{"id":"3ebed33a-a80c-4888-b3d3-2677e87c25e7"}
```

## Get component

```bash
$ curl -sL \
https://status.k8s.scs.community/components/3ebed33a-a80c-4888-b3d3-2677e87c25e7

{"data":{"activelyAffectedBy":[{"reference":"09f471bb-b0af-4528-a021-f23f31e2d1c9","severity":90,"type":"52d178d2-0fe9-4654-834b-42502283454d"}],"displayName":"Test-Component","id":"3ebed33a-a80c-4888-b3d3-2677e87c25e7"}}
```

## List incidents

```bash
$ curl -sL \
'http://localhost:3000/incidents?start=2024-04-01T10%3A10%3A10.010Z&end=2024-04-30T10%3A10%3A10.010Z'

{"data":[{"affects":[{"reference":"3ebed33a-a80c-4888-b3d3-2677e87c25e7","type":"52d178d2-0fe9-4654-834b-42502283454d"}],"beganAt":"2024-04-03T08:00:00+02:00","description":"A test incident.","displayName":"Test-Incident","endedAt":null,"id":"09f471bb-b0af-4528-a021-f23f31e2d1c9","phase":{"generation":0,"order":0},"updates":[]}]}
```

## Create incident

```bash
$ curl -sL \
-X POST \
-H 'Authorization: Bearer <your-id-token>' \
-H 'Content-Type: application/json' \
-d '{"affects":[{"reference":"3ebed33a-a80c-4888-b3d3-2677e87c25e7","severity":90,"type":"52d178d2-0fe9-4654-834b-42502283454d"}],"beganAt":"2024-04-03T06:00:00.000Z","description":"A test incident.","displayName":"Test-Incident","phase":{}}' \
https://status.k8s.scs.community/incidents

{"id":"09f471bb-b0af-4528-a021-f23f31e2d1c9"}
```

## Get incident

```bash
$ curl -sL \
https://status.k8s.scs.community/incidents/09f471bb-b0af-4528-a021-f23f31e2d1c9

{"data":{"affects":[{"reference":"3ebed33a-a80c-4888-b3d3-2677e87c25e7","type":"52d178d2-0fe9-4654-834b-42502283454d"}],"beganAt":"2024-04-03T08:00:00+02:00","description":"A test incident.","displayName":"Test-Incident","endedAt":null,"id":"09f471bb-b0af-4528-a021-f23f31e2d1c9","phase":{"generation":0,"order":0},"updates":[]}}
```

## List incident update

```bash
$ curl -sL \
https://status.k8s.scs.community/incidents/09f471bb-b0af-4528-a021-f23f31e2d1c9/updates

{"data":[{"createdAt":"2024-04-03T08:15:00+02:00","description":"Example update for test incident","displayName":"Example Update","order":0}]}
```

## Create incident update

```bash
$ curl -sL \
-X POST \
-H 'Authorization: Bearer <your-id-token>' \
-H 'Content-Type: application/json' \
-d '{"createdAt":"2024-04-03T06:15:00.000Z","description":"Example update for test incident","displayName":"Example Update"}' \
https://status.k8s.scs.community/incidents/09f471bb-b0af-4528-a021-f23f31e2d1c9/updates

{"order": 0}
```

## Get incident update

```bash
$ curl -sL \
https://status.k8s.scs.community/incidents/09f471bb-b0af-4528-a021-f23f31e2d1c9/updates/0

{"data":{"createdAt":"2024-04-03T08:15:00+02:00","description":"Example update for test incident","displayName":"Example Update","order":0}}
```
