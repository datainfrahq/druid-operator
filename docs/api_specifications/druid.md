<h1>Druid API reference</h1>
<p>Packages:</p>
<ul class="simple">
<li>
<a href="#druid.apache.org%2fv1alpha1">druid.apache.org/v1alpha1</a>
</li>
</ul>
<h2 id="druid.apache.org/v1alpha1">druid.apache.org/v1alpha1</h2>
Resource Types:
<ul class="simple"></ul>
<h3 id="druid.apache.org/v1alpha1.AdditionalContainer">AdditionalContainer
</h3>
<p>
(<em>Appears on:</em>
<a href="#druid.apache.org/v1alpha1.DruidSpec">DruidSpec</a>)
</p>
<p>AdditionalContainer defines the additional sidecar container</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>runAsInit</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>image</code><br>
<em>
string
</em>
</td>
<td>
<p>This is the image for the additional container to run.</p>
</td>
</tr>
<tr>
<td>
<code>containerName</code><br>
<em>
string
</em>
</td>
<td>
<p>This is the name of the additional container.</p>
</td>
</tr>
<tr>
<td>
<code>command</code><br>
<em>
[]string
</em>
</td>
<td>
<p>This is the command for the additional container to run.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>If not present, will be taken from top level spec</p>
</td>
</tr>
<tr>
<td>
<code>args</code><br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Argument to call the command</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#securitycontext-v1-core">
Kubernetes core/v1.SecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ContainerSecurityContext. If not present, will be taken from top level pod</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CPU/Memory Resources</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Volumes etc for the Druid pods</p>
</td>
</tr>
<tr>
<td>
<code>env</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Environment variables for the Additional Container</p>
</td>
</tr>
<tr>
<td>
<code>envFrom</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#envfromsource-v1-core">
[]Kubernetes core/v1.EnvFromSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Extra environment variables</p>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="druid.apache.org/v1alpha1.DeepStorageSpec">DeepStorageSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#druid.apache.org/v1alpha1.DruidSpec">DruidSpec</a>)
</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
encoding/json.RawMessage
</em>
</td>
<td>
<br/>
<br/>
<table>
</table>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="druid.apache.org/v1alpha1.Druid">Druid
</h3>
<p>Druid is the Schema for the druids API</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.DruidSpec">
DruidSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>ignored</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ignored is now deprecated API. In order to avoid reconciliation of objects use the
druid.apache.org/ignored: &ldquo;true&rdquo; annotation</p>
</td>
</tr>
<tr>
<td>
<code>common.runtime.properties</code><br>
<em>
string
</em>
</td>
<td>
<p>common.runtime.properties contents</p>
</td>
</tr>
<tr>
<td>
<code>extraCommonConfig</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#*k8s.io/api/core/v1.objectreference--">
[]*k8s.io/api/core/v1.ObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>References to ConfigMaps holding more files to mount to the CommonConfigMountPath.</p>
</td>
</tr>
<tr>
<td>
<code>forceDeleteStsPodOnError</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>scalePvcSts</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>ScalePvcSts, defaults to false. When enabled, operator will allow volume expansion of sts and pvc&rsquo;s.</p>
</td>
</tr>
<tr>
<td>
<code>commonConfigMountPath</code><br>
<em>
string
</em>
</td>
<td>
<p>In-container directory to mount with common.runtime.properties</p>
</td>
</tr>
<tr>
<td>
<code>disablePVCDeletionFinalizer</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Default is set to false, pvc shall be deleted on deletion of CR</p>
</td>
</tr>
<tr>
<td>
<code>deleteOrphanPvc</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Default is set to true, orphaned ( unmounted pvc&rsquo;s ) shall be cleaned up by the operator.</p>
</td>
</tr>
<tr>
<td>
<code>startScript</code><br>
<em>
string
</em>
</td>
<td>
<p>Path to druid start script to be run on container start</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Required here or at nodeSpec level</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccount</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ServiceAccount for the druid cluster</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>imagePullSecrets for private registries</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>env</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Environment variables for druid containers</p>
</td>
</tr>
<tr>
<td>
<code>envFrom</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#envfromsource-v1-core">
[]Kubernetes core/v1.EnvFromSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Extra environment variables</p>
</td>
</tr>
<tr>
<td>
<code>jvm.options</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>jvm options for druid jvm processes</p>
</td>
</tr>
<tr>
<td>
<code>log4j.config</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>log4j config contents</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>druid pods pod-security-context</p>
</td>
</tr>
<tr>
<td>
<code>containerSecurityContext</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#securitycontext-v1-core">
Kubernetes core/v1.SecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>druid pods container-security-context</p>
</td>
</tr>
<tr>
<td>
<code>volumeClaimTemplates</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#persistentvolumeclaim-v1-core">
[]Kubernetes core/v1.PersistentVolumeClaim
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumes etc for the Druid pods</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>podAnnotations</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Custom annotations to be populated in Druid pods</p>
</td>
</tr>
<tr>
<td>
<code>podManagementPolicy</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#podmanagementpolicytype-v1-apps">
Kubernetes apps/v1.PodManagementPolicyType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>By default, it is set to &ldquo;parallel&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>podLabels</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Custom labels to be populated in Druid pods</p>
</td>
</tr>
<tr>
<td>
<code>updateStrategy</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#statefulsetupdatestrategy-v1-apps">
Kubernetes apps/v1.StatefulSetUpdateStrategy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>livenessProbe</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Port is set to druid.port if not specified with httpGet handler</p>
</td>
</tr>
<tr>
<td>
<code>readinessProbe</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Port is set to druid.port if not specified with httpGet handler</p>
</td>
</tr>
<tr>
<td>
<code>startUpProbe</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StartupProbe for nodeSpec</p>
</td>
</tr>
<tr>
<td>
<code>services</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#service-v1-core">
[]Kubernetes core/v1.Service
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>k8s service resources to be created for each Druid statefulsets</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Node selector to be used by Druid statefulsets</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Toleration to be used in order to run Druid on nodes tainted</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Affinity to be used to for enabling node, pod affinity and anti-affinity</p>
</td>
</tr>
<tr>
<td>
<code>nodes</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.DruidNodeSpec">
map[string]druid-operator/apis/druid/v1alpha1.DruidNodeSpec
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>additionalContainer</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.AdditionalContainer">
[]AdditionalContainer
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Operator deploys the sidecar container based on these properties. Sidecar will be deployed for all the Druid pods.</p>
</td>
</tr>
<tr>
<td>
<code>rollingDeploy</code><br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>zookeeper</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.ZookeeperSpec">
ZookeeperSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>futuristic stuff to make Druid dependency setup extensible from within Druid operator
ignore for now.</p>
</td>
</tr>
<tr>
<td>
<code>metadataStore</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.MetadataStoreSpec">
MetadataStoreSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>deepStorage</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.DeepStorageSpec">
DeepStorageSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>metricDimensions.json</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Custom Dimension Map Path for statsd emitter</p>
</td>
</tr>
<tr>
<td>
<code>hdfs-site.xml</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>HDFS common config</p>
</td>
</tr>
<tr>
<td>
<code>core-site.xml</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.DruidClusterStatus">
DruidClusterStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="druid.apache.org/v1alpha1.DruidClusterStatus">DruidClusterStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#druid.apache.org/v1alpha1.Druid">Druid</a>)
</p>
<p>DruidStatus defines the observed state of Druid</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>druidNodeStatus</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.DruidNodeTypeStatus">
DruidNodeTypeStatus
</a>
</em>
</td>
<td>
<p>INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
Important: Run &ldquo;make&rdquo; to regenerate code after modifying this file</p>
</td>
</tr>
<tr>
<td>
<code>statefulSets</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>deployments</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>services</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>configMaps</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>podDisruptionBudgets</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>ingress</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>hpAutoscalers</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>pods</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>persistentVolumeClaims</code><br>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="druid.apache.org/v1alpha1.DruidNodeConditionType">DruidNodeConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em>
<a href="#druid.apache.org/v1alpha1.DruidNodeTypeStatus">DruidNodeTypeStatus</a>)
</p>
<h3 id="druid.apache.org/v1alpha1.DruidNodeSpec">DruidNodeSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#druid.apache.org/v1alpha1.DruidSpec">DruidSpec</a>)
</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>nodeType</code><br>
<em>
string
</em>
</td>
<td>
<p>Druid node type</p>
</td>
</tr>
<tr>
<td>
<code>druid.port</code><br>
<em>
int32
</em>
</td>
<td>
<p>Port used by Druid Process</p>
</td>
</tr>
<tr>
<td>
<code>kind</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Defaults to statefulsets.
Note: volumeClaimTemplates are ignored when kind=Deployment</p>
</td>
</tr>
<tr>
<td>
<code>replicas</code><br>
<em>
int32
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>podLabels</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>podDisruptionBudgetSpec</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#poddisruptionbudgetspec-v1-policy">
Kubernetes policy/v1.PodDisruptionBudgetSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>runtime.properties</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>jvm.options</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>This overrides JvmOptions at top level</p>
</td>
</tr>
<tr>
<td>
<code>extra.jvm.options</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>This appends extra jvm options to JvmOptions field</p>
</td>
</tr>
<tr>
<td>
<code>log4j.config</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>This overrides Log4jConfig at top level</p>
</td>
</tr>
<tr>
<td>
<code>nodeConfigMountPath</code><br>
<em>
string
</em>
</td>
<td>
<p>in-container directory to mount with runtime.properties, jvm.config, log4j2.xml files</p>
</td>
</tr>
<tr>
<td>
<code>services</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#service-v1-core">
[]Kubernetes core/v1.Service
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overrides services at top level</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Toleration to be used in order to run Druid on nodes tainted</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Affinity to be used to for enabling node, pod affinity and anti-affinity</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Node selector to be used by Druid statefulsets</p>
</td>
</tr>
<tr>
<td>
<code>terminationGracePeriodSeconds</code><br>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>ports</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#containerport-v1-core">
[]Kubernetes core/v1.ContainerPort
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Extra ports to be added to pod spec</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overrides image from top level, Required if no image specified at top level</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overrides imagePullSecrets from top level</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overrides imagePullPolicy from top level</p>
</td>
</tr>
<tr>
<td>
<code>env</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Extra environment variables</p>
</td>
</tr>
<tr>
<td>
<code>envFrom</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#envfromsource-v1-core">
[]Kubernetes core/v1.EnvFromSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Extra environment variables</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcerequirements-v1-core">
Kubernetes core/v1.ResourceRequirements
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CPU/Memory Resources</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overrides securityContext at top level</p>
</td>
</tr>
<tr>
<td>
<code>containerSecurityContext</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#securitycontext-v1-core">
Kubernetes core/v1.SecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Druid pods container-security-context</p>
</td>
</tr>
<tr>
<td>
<code>podAnnotations</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Custom annotations to be populated in Druid pods</p>
</td>
</tr>
<tr>
<td>
<code>podManagementPolicy</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#podmanagementpolicytype-v1-apps">
Kubernetes apps/v1.PodManagementPolicyType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>By default, it is set to &ldquo;parallel&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>maxSurge</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>maxSurge for deployment object, only applicable if kind=Deployment, by default set to 25%</p>
</td>
</tr>
<tr>
<td>
<code>maxUnavailable</code><br>
<em>
int32
</em>
</td>
<td>
<em>(Optional)</em>
<p>maxUnavailable for deployment object, only applicable if kind=Deployment, by default set to 25%</p>
</td>
</tr>
<tr>
<td>
<code>updateStrategy</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#statefulsetupdatestrategy-v1-apps">
Kubernetes apps/v1.StatefulSetUpdateStrategy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>livenessProbe</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>readinessProbe</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>startUpProbes</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StartupProbe for nodeSpec</p>
</td>
</tr>
<tr>
<td>
<code>ingressAnnotations</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ingress Annoatations to be populated in ingress spec</p>
</td>
</tr>
<tr>
<td>
<code>ingress</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#ingressspec-v1-networking">
Kubernetes networking/v1.IngressSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ingress Spec</p>
</td>
</tr>
<tr>
<td>
<code>persistentVolumeClaim</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#persistentvolumeclaim-v1-core">
[]Kubernetes core/v1.PersistentVolumeClaim
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>lifecycle</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#lifecycle-v1-core">
Kubernetes core/v1.Lifecycle
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>hpAutoscaler</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#horizontalpodautoscalerspec-v2-autoscaling">
Kubernetes autoscaling/v2.HorizontalPodAutoscalerSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>topologySpreadConstraints</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#topologyspreadconstraint-v1-core">
[]Kubernetes core/v1.TopologySpreadConstraint
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>volumeClaimTemplates</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#persistentvolumeclaim-v1-core">
[]Kubernetes core/v1.PersistentVolumeClaim
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="druid.apache.org/v1alpha1.DruidNodeTypeStatus">DruidNodeTypeStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#druid.apache.org/v1alpha1.DruidClusterStatus">DruidClusterStatus</a>)
</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>druidNode</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>druidNodeConditionStatus</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#conditionstatus-v1-core">
Kubernetes core/v1.ConditionStatus
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>druidNodeConditionType</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.DruidNodeConditionType">
DruidNodeConditionType
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>reason</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="druid.apache.org/v1alpha1.DruidSpec">DruidSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#druid.apache.org/v1alpha1.Druid">Druid</a>)
</p>
<p>DruidSpec defines the desired state of Druid</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>ignored</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ignored is now deprecated API. In order to avoid reconciliation of objects use the
druid.apache.org/ignored: &ldquo;true&rdquo; annotation</p>
</td>
</tr>
<tr>
<td>
<code>common.runtime.properties</code><br>
<em>
string
</em>
</td>
<td>
<p>common.runtime.properties contents</p>
</td>
</tr>
<tr>
<td>
<code>extraCommonConfig</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#*k8s.io/api/core/v1.objectreference--">
[]*k8s.io/api/core/v1.ObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>References to ConfigMaps holding more files to mount to the CommonConfigMountPath.</p>
</td>
</tr>
<tr>
<td>
<code>forceDeleteStsPodOnError</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>scalePvcSts</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>ScalePvcSts, defaults to false. When enabled, operator will allow volume expansion of sts and pvc&rsquo;s.</p>
</td>
</tr>
<tr>
<td>
<code>commonConfigMountPath</code><br>
<em>
string
</em>
</td>
<td>
<p>In-container directory to mount with common.runtime.properties</p>
</td>
</tr>
<tr>
<td>
<code>disablePVCDeletionFinalizer</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Default is set to false, pvc shall be deleted on deletion of CR</p>
</td>
</tr>
<tr>
<td>
<code>deleteOrphanPvc</code><br>
<em>
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>Default is set to true, orphaned ( unmounted pvc&rsquo;s ) shall be cleaned up by the operator.</p>
</td>
</tr>
<tr>
<td>
<code>startScript</code><br>
<em>
string
</em>
</td>
<td>
<p>Path to druid start script to be run on container start</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Required here or at nodeSpec level</p>
</td>
</tr>
<tr>
<td>
<code>serviceAccount</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>ServiceAccount for the druid cluster</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecrets</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>imagePullSecrets for private registries</p>
</td>
</tr>
<tr>
<td>
<code>imagePullPolicy</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#pullpolicy-v1-core">
Kubernetes core/v1.PullPolicy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>env</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#envvar-v1-core">
[]Kubernetes core/v1.EnvVar
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Environment variables for druid containers</p>
</td>
</tr>
<tr>
<td>
<code>envFrom</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#envfromsource-v1-core">
[]Kubernetes core/v1.EnvFromSource
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Extra environment variables</p>
</td>
</tr>
<tr>
<td>
<code>jvm.options</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>jvm options for druid jvm processes</p>
</td>
</tr>
<tr>
<td>
<code>log4j.config</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>log4j config contents</p>
</td>
</tr>
<tr>
<td>
<code>securityContext</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#podsecuritycontext-v1-core">
Kubernetes core/v1.PodSecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>druid pods pod-security-context</p>
</td>
</tr>
<tr>
<td>
<code>containerSecurityContext</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#securitycontext-v1-core">
Kubernetes core/v1.SecurityContext
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>druid pods container-security-context</p>
</td>
</tr>
<tr>
<td>
<code>volumeClaimTemplates</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#persistentvolumeclaim-v1-core">
[]Kubernetes core/v1.PersistentVolumeClaim
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>volumes etc for the Druid pods</p>
</td>
</tr>
<tr>
<td>
<code>volumeMounts</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#volumemount-v1-core">
[]Kubernetes core/v1.VolumeMount
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#volume-v1-core">
[]Kubernetes core/v1.Volume
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>podAnnotations</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Custom annotations to be populated in Druid pods</p>
</td>
</tr>
<tr>
<td>
<code>podManagementPolicy</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#podmanagementpolicytype-v1-apps">
Kubernetes apps/v1.PodManagementPolicyType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>By default, it is set to &ldquo;parallel&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>podLabels</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Custom labels to be populated in Druid pods</p>
</td>
</tr>
<tr>
<td>
<code>updateStrategy</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#statefulsetupdatestrategy-v1-apps">
Kubernetes apps/v1.StatefulSetUpdateStrategy
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>livenessProbe</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Port is set to druid.port if not specified with httpGet handler</p>
</td>
</tr>
<tr>
<td>
<code>readinessProbe</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Port is set to druid.port if not specified with httpGet handler</p>
</td>
</tr>
<tr>
<td>
<code>startUpProbe</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#probe-v1-core">
Kubernetes core/v1.Probe
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>StartupProbe for nodeSpec</p>
</td>
</tr>
<tr>
<td>
<code>services</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#service-v1-core">
[]Kubernetes core/v1.Service
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>k8s service resources to be created for each Druid statefulsets</p>
</td>
</tr>
<tr>
<td>
<code>nodeSelector</code><br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Node selector to be used by Druid statefulsets</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#toleration-v1-core">
[]Kubernetes core/v1.Toleration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Toleration to be used in order to run Druid on nodes tainted</p>
</td>
</tr>
<tr>
<td>
<code>affinity</code><br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#affinity-v1-core">
Kubernetes core/v1.Affinity
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Affinity to be used to for enabling node, pod affinity and anti-affinity</p>
</td>
</tr>
<tr>
<td>
<code>nodes</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.DruidNodeSpec">
map[string]druid-operator/apis/druid/v1alpha1.DruidNodeSpec
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>additionalContainer</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.AdditionalContainer">
[]AdditionalContainer
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Operator deploys the sidecar container based on these properties. Sidecar will be deployed for all the Druid pods.</p>
</td>
</tr>
<tr>
<td>
<code>rollingDeploy</code><br>
<em>
bool
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>zookeeper</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.ZookeeperSpec">
ZookeeperSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>futuristic stuff to make Druid dependency setup extensible from within Druid operator
ignore for now.</p>
</td>
</tr>
<tr>
<td>
<code>metadataStore</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.MetadataStoreSpec">
MetadataStoreSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>deepStorage</code><br>
<em>
<a href="#druid.apache.org/v1alpha1.DeepStorageSpec">
DeepStorageSpec
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>metricDimensions.json</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Custom Dimension Map Path for statsd emitter</p>
</td>
</tr>
<tr>
<td>
<code>hdfs-site.xml</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>HDFS common config</p>
</td>
</tr>
<tr>
<td>
<code>core-site.xml</code><br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="druid.apache.org/v1alpha1.MetadataStoreSpec">MetadataStoreSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#druid.apache.org/v1alpha1.DruidSpec">DruidSpec</a>)
</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
encoding/json.RawMessage
</em>
</td>
<td>
<br/>
<br/>
<table>
</table>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<h3 id="druid.apache.org/v1alpha1.ZookeeperSpec">ZookeeperSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#druid.apache.org/v1alpha1.DruidSpec">DruidSpec</a>)
</p>
<div class="md-typeset__scrollwrap">
<div class="md-typeset__table">
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code><br>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>spec</code><br>
<em>
encoding/json.RawMessage
</em>
</td>
<td>
<br/>
<br/>
<table>
</table>
</td>
</tr>
</tbody>
</table>
</div>
</div>
<div class="admonition note">
<p class="last">This page was automatically generated with <code>gen-crd-api-reference-docs</code></p>
</div>
