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
During upgradeS, if THE StatefulSet is set to `OrderedReady` - the StatefulSet controller will not recover from 
crash-loopback state. The issues is referenced [here](https://github.com/kubernetes/kubernetes/issues/67250). 
Documentation reference: [doc](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#forced-rollback)
The operator solves this by using the `forceDeleteStsPodOnError` key, the operator will delete the sts pod if its in 
crash-loopback state.  
Example scenario: During upgrade, user rolls out a faulty configuration causing the historical pod going in crashing 
state. Then, the user rolls out a valid configuration - the new configuration will not be applied unless user manually 
delete the pods. To solve this scenario, the operator will delete the pod automatically without user intervention.  

```
NOTE: User must be aware of this feature, there might be cases where crash-loopback might be caused due probe failure, 
fault image etc, the operator will keep on deleting on each re-concile loop. Default Behavior is True.
```

## Horizontal Scaling of Druid Pods
The operator supports the `HPA autosaling/v2` specification in the `nodeSpec` for druid nodes. In case an HPA deployed, 
the HPA controller maintains the replica count/state for the particular workload referenced.  
Refer to `examples.md` for HPA configuration. 

```
NOTE: This option in currently prefered to scale only brokers using HPA. In order to scale Middle Managers with HPA, 
its recommended not to use HPA. Refer to these discussions which have adderessed the issues in details:
```
1. <https://github.com/apache/druid/issues/8801#issuecomment-664020630>
2. <https://github.com/apache/druid/issues/8801#issuecomment-664648399>

## Volume Expansion of Druid Pods Running As StatefulSets
```
NOTE: This feature has been tested only on cloud environments and storage classes which have supported volume expansion.
This feature uses cascade=orphan strategy to make sure that only the StatefulSet is deleted and recreated and pods 
are not deleted.
```
Druid Nodes (specifically historical nodes) run as StatefulSets. Each StatefulSet replica has a PVC attached. The 
`NodeSpec` in Druid CR has the key `volumeClaimTemplates` where users can define the PVC's storage class as well 
as size. Currently, in Kubernetes, in case a user wants to increase the size in the node, the StatefulSets cannot 
be directly updated. The Druid operator can perform a seamless update of the StatefulSet, and patch the 
PVCs with the desired size defined in the druid CR. Behind the scenes, the operator performs a cascade deletion of the 
StatefulSet, and patches the PVC. Cascade deletion has no affect to the pods running (queries are served and no 
downtime is experienced).  
While enabling this feature, the operator will check if volume expansion is supported in the storage class mentioned 
in the druid CR, only then will it perform expansion. 
This feature is disabled by default. To enable it set `scalePvcSts: true` in the Druid CR.
By default, this feature is disabled.

```
IMPORTANT: Shrinkage of pvc's isnt supported - desiredSize cannot be less than currentSize as well as counts. 
```

## Add Additional Containers to Druid Pods
The operator supports adding additional containers to run along with the druid pods. This helps support co-located, 
co-managed helper processes for the primary druid application. This can be used for init containers, sidecars, 
proxies etc.  
To enable this features users just need to add new containers to the `AdditionalContainers` in the Druid spec API.
```
NOTE: This is scoped at cluster scope only, which means that additional container will be common to all the nodes. 
```

## Default Yet Configurable Probes
The operator create the Deployments and StatefulSets with a default set of probes for each druid components.
These probes can be overriden by adding one of the probes in the `DruidSpec` (global) or under the
`NodeSpec` (component-scope). 
All the probes definitions are documented bellow:

<details>

<summary>Coordinator, Overlord, MiddleManager, Router and Indexer probes</summary>

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
