# Features

- [Deny List in Operator](#deny-list-in-operator)
- [Reconcile Time in Operator](#reconcile-time-in-operator)
- [Finalizer in Druid CR](#finalizer-in-druid-cr)
- [Deletetion of Orphan PVC's](#deletetion-of-orphan-pvcs)
- [Rolling Deploy](#rolling-deploy)
- [Force Delete of Sts Pods](#force-delete-of-sts-pods)
- [Scaling of Druid Nodes](#scaling-of-druid-nodes)
- [Volume Expansion of Druid Nodes Running As StatefulSets](#volume-expansion-of-druid-nodes-running-as-statefulsets)
- [Add Additional Containers in Druid Nodes](#add-additional-containers-in-druid-nodes)
- [Setup default probe by default](#setup-default-probe-by-default)


## Deny List in Operator
There may be use cases where we want the operator to watch all namespaces except a few 
(might be due to security, testing flexibility, etc. reasons).  
Druid operator supports such cases - in the chart, edit `env.DENY_LIST` to be a comma-seperated list.  
For example: "default,kube-system"

## Reconcile Time in Operator
As per operator pattern, the druid operator reconciles every 10s (default reconciliation time) to make sure 
the desired state (in that case, the druid CR's spec) is in sync with the current state.  
The reconciliation time can be adjusted - in the chart, add `env.RECONCILE_WAIT` to be a duration
in seconds.  
Examples: "10s", "30s", "120s"

## Finalizer in Druid CR
The Druid operator supports provisioning of StatefulSets and Deployments. When a StatefulSet is created, 
a PVC is created along. When the Druid CR is deleted, the StatefulSet controller does not delete the PVC's 
associated with it.  
In case the PVC data is important and you wish to reclaim it, you can enable: `DisablePVCDeletionFinalizer: true`
in the Druid CR.  
The default behavior is to trigger finalizers and pre-delete hooks that will be executed. They will first clean up the 
StatefulSet and then the PVCs referenced to it. That means that after a 
deletion of a Druid CR, any PVCs provisioned by a StatefulSet will be deleted.

## Deletion of Orphan PVCs
There are some use-cases (the most popular is horizontal auto-scaling) where a StatefulSet scales down. In that case,
the statefulSet will terminate its owned pods but nit their attached PVCs which left orphaned and unused.  
The operator support the ability to auto delete these PVCs. This can be enabled by setting `deleteOrphanPvc: true`.

## Rolling Deploy
The operator supports Apache Druid's recommended rolling updates. It will do incremental updates in the order
specified in Druid's [documentation](https://druid.apache.org/docs/latest/operations/rolling-updates.html).  
In case any of the node goes in pending/crashing state during an update, the operator halts the update and does
not continue with the update - this will require a manual intervention.  
Default updates are done in parallel. Since cluster creation does not require a rolling update, they will be done
in parallel anyway. To enable this feature, set `rollingDeploy: true` in the Druid CR.

## Force Delete of Sts Pods

- During upgrade if sts is set to ordered ready, the sts controller will not recover from crashloopback state. The issues is referenced [here](https://github.com/kubernetes/kubernetes/issues/67250), and here's a reference [doc](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#forced-rollback)
- How operator solves this is using the ```forceDeleteStsPodOnError``` key, the operator will delete the sts pod if its in crashloopback state. Example Scenario: During upgrade, user rolls out a faulty configuration causing the historical pod going in crashing state, user rolls out a valid configuration, the new configuration will not be applied unless user manual delete pods, so solve this scenario operator shall delete the pod automatically without user intervention.
- ```NOTE: User must be aware of this feature, there might be cases where crashloopback might be caused due probe failure, fault image etc, the operator shall keep on deleting on each re-concile loop. Default Behavior is True ```

## Scaling of Druid Nodes

- Operator supports ```HPA autosaling/v2``` Spec in the nodeSpec for druid nodes. In case HPA deployed, HPA controller maintains the replica count/state for the particular statefulset referenced.  Refer to ```examples.md``` for HPA configuration. 
- ```NOTE: Prefered to scale only brokers using HPA.```
- In order to scale MM with HPA, its recommended not to use HPA. Refer to these discussions which have adderessed the issues in details.

1. <https://github.com/apache/druid/issues/8801#issuecomment-664020630>
2. <https://github.com/apache/druid/issues/8801#issuecomment-664648399>

## Volume Expansion of Druid Nodes Running As StatefulSets

```NOTE: This feature has been tested only on cloud environments and storage classes which have supported volume expansion. This feature uses cascade=orphan strategy to make sure only Stateful is deleted and recreated and pods are not deleted.```

- Druid Nodes specifically historicals run as statefulsets. Each statefulset replica has a pvc attached.
- NodeSpec in druid CR has key ```volumeClaimTemplates``` where users can define the pvc's storage class as well as size.
- In case a user wants to increase size in the node, the statefulsets cannot be directly updated.
- Druid Operator behind the scenes performs seamless update of the statefulset, plus patch the pvc's with desired size defined in the druid CR.
- Druid operator shall perform a cascade deletion of the sts, and shall patch the pvc. Cascade deletion has no affect to the pods running, queries are served and no downtime is experienced.
- While enabling this feature, druid operator will check if volume expansion is supported in the storage class mentioned in the druid CR, only then will it perform expansion.
- Shrinkage of pvc's isnt supported, **desiredSize cannot be less than currentSize as well as counts**.
- To enable this feature ```scalePvcSts``` needs to be enabled to ```true```.
- By default, this feature is disabled.

## Add Additional Containers in Druid Nodes

- The Druid operator supports additional containers to run along with the druid services. This helps support co-located, co-managed helper processes for the primary druid application
- This can be used for init containers or sidecars or proxies etc.
- To enable this features users just need to add a new container to the container list.
- This is scoped at cluster scope only, which means that additional container will be common to all the nodes.
- This can be used for init containers or sidecars or proxies etc. 
- To enable this features users just need to add a new container to the container list 
- This is scoped at cluster scope only, which means that additional container will be common to all the nodes

## Setup default probe by default

The operator create deployments and statefullset with a default set of probes for each druid components.
Theses probes are overrided if you specify a global or specific probe in the druid resource.
All the probes definitions are documented bellow:

<details>

<summary>Coordinator, Overlord, Middlemanager, Router and Indexer probes</summary>

```yaml
  livenessProbe:
    failureThreshold: 10
    httpGet:
      path: /status/health
      port: $druid.port
    initialDelaySeconds: 5
    periodSeconds: 10
    successThreshold: 1
    timeoutSeconds: 5
  readinessProbe:
    failureThreshold: 10
    httpGet:
      path: /status/health
      port: $druid.port
    initialDelaySeconds: 5
    periodSeconds: 10
    successThreshold: 1
    timeoutSeconds: 5
  startupProbe:
    failureThreshold: 10
    httpGet:
      path: /status/health
      port: $druid.port
    initialDelaySeconds: 5
    periodSeconds: 10
    successThreshold: 1
    timeoutSeconds: 5
```

</details>

<details>

<summary>Broker probes </summary>

  ```yaml
      livenessProbe:
        failureThreshold: 10
        httpGet:
          path: /status/health
          port: $druid.port
        initialDelaySeconds: 5
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 5
      readinessProbe:
        failureThreshold: 20
        httpGet:
          path: /druid/broker/v1/readiness
          port: $druid.port
        initialDelaySeconds: 5
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 5
      startupProbe:
        failureThreshold: 20
        httpGet:
          path: /druid/broker/v1/readiness
          port: $druid.port
        initialDelaySeconds: 5
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 5
  ```
</details>

<details>

<summary>Historical probes</summary>

```yaml
  livenessProbe:
    failureThreshold: 10
    httpGet:
      path: /status/health
      port: $druid.port
    initialDelaySeconds: 5
    periodSeconds: 10
    successThreshold: 1
    timeoutSeconds: 5
  readinessProbe:
    failureThreshold: 20
    httpGet:
      path: /druid/historical/v1/loadstatus
      port: $druid.port
    initialDelaySeconds: 5
    periodSeconds: 10
    successThreshold: 1
    timeoutSeconds: 5
  startupProbe:
    failureThreshold: 20
    httpGet:
      path: /druid/historical/v1/loadstatus
      port: $druid.port
    initialDelaySeconds: 180
    periodSeconds: 30
    successThreshold: 1
    timeoutSeconds: 10
```

</details>
