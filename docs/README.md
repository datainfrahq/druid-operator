# Overview

Druid Operator is a Kubernetes controller that manages the lifecycle of [Apache Druid](https://druid.apache.org/) clusters.
The operator simplifies the management of Druid clusters with its custom logic that is configurable via custom API
(Kubernetes CRD).

## Druid Operator Documentation

* [Getting Started](./getting_started.md)
* API Specifications
  * [Druid API](./api_specifications/druid.md)
* [Feature Supported by Druid Operator](./features.md)
* [Example Specs](./example_specs.md)
* [Migration To Kubebuilder V3 in the Upcoming Version](./kubebuilder_v3_migration.md)
* [Developer Documentation](./dev_doc.md)
---

:warning: You won't find any documentation about druid itself in this repository.
If you need details about how to architecture your druid cluster you can consult theses documentations:

* [Druid introduction](<https://druid.apache.org/docs/latest/design/index.html>)
* [Druid architecture](https://druid.apache.org/docs/latest/design/architecture.html)
* [Druid configuration reference](https://druid.apache.org/docs/latest/configuration/index.html)
