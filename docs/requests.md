# Requests

As defined by the [OpenAPI spec](https://github.com/SovereignCloudStack/status-page-openapi) and [status page OpenAPI decision](https://github.com/SovereignCloudStack/standards/blob/main/Standards/scs-0402-v1-status-page-openapi-spec-decision.md) the general API objects are used as request bodies and respponses to generalize data structures. Not all object fields are handled by all requests, some are read only and some are write only. `GET` request wrap their return in a object field called `data`.

Please refer to [example requests](example-requests.md) to see these request in action.

## Phases

Phases are always handled as lists, so `GET` as well as `POST` operations on phases always require the full list. When geting the phase list, it's accompanied be a generation annotation.

```json
{
  "generation": 1, // omitted on POST and PATCH
  "phases": [
    "Phase 1",
    "Phase 2",
    "Pahse 3"
  ]
}
```

## Impact types

Requesting (`GET`) an impact type, will return all fields, while `POST` and `PATCH` operations ommit the `id` field.

```json
{
  "id": "UUID", // ommit on POST and PATCH
  "description": "Desription of the impact type.",
  "displayName": "Name"
}
```

## Severities

For all request types (`GET`, `POST`, `PATCH`), all fields of the severity are handled.

```json
{
  "displayName": "string",
  "value": 100
}
```

As `displayName` is the identifier it must be unique, even when modified by `PATCH`

## Components

When `GET`ing a component, all fields can be expected to be filled, while requests for `POST` (creation) and `PATCH` operations only handle certain fields.

```json
{
  "id": "UUID", // ommited on POST and PATCH
  "activelyAffectedBy": [ // ommited on POST and PATCH
    {
      "reference": "Incident-UUID",
      "severity": 100,
      "type": "ImpactType-UUID"
    }
  ],
  "displayName": "Name",
  "labels": {
    "key": "value",
  }
}
```

## Incidents

It is expected that incidents are the most used API object and have the most data to transmit.

```json
{
  "id": "UUID", // ommited on POST and PATCH
  "affects": [
    {
      "reference": "Component-UUID",
      "severity": 100,
      "type": "ImpactType-UUID"
    }
  ],
  "beganAt": "2024-01-01T06:00:00.000Z",
  "description": "Description of the incident.",
  "displayName": "Name",
  "endedAt": "2024-01-01T08:00:00.000Z", // or null, when incident is still ongoing.
  "phase": {
    "generation": 1,
    "order": 1
  },
  "updates": [ // ommited on POST and PATCH
    0,
    1,
    2
  ]
}
```

The `affects` field can, in theory, have as many entries as there are components. The `updates` field is an ongoing list of updates. All changes to `phase` should correlate to an update.

When performing `POST` or `PATCH` operations on incidents the `affects` field is of utmost importance, as it creates the **impact**. Only when referencing a component to an incident via the `affects` field, an impact is created, that can be retrieved via the affected component.

## Incident update

Whenever an incident changes, an update should be issued. When doing a `GET` request, the `order` field is filled, updates should be displayed in ascending order.

```json
{
  "order": 0, // ommited on POST and PATCH
  "createdAt": "2024-01-01T06:15:00.000Z",
  "description": "Description of the update.",
  "displayName": "Name"
}
```
