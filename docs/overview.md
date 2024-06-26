# Overview

The `status-page-api` repository strives to provide an example implementation of the concepts being outlined by [`status-page-openapi`](https://github.com/SovereignCloudStack/status-page-openapi).

For the implementation [`go`](https://go.dev/) was used, as rational see the [decision record](https://github.com/SovereignCloudStack/standards/blob/main/Standards/scs-0401-v1-status-page-reference-implementation-decision.md#programming-language).

Code under the `pkg/` directory, is considered as public example or even library code to implement your own API server, while code in `internal/` is implementation specific to this application, but can serve as example too.
