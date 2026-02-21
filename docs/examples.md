<!--
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
-->
## Configure Hot/Cold for Historical Pods
```yaml
...
  nodes:
    hot:
      druid.port: 8083
      env:
      - name: DRUID_XMS
        value: 2g
      - name: DRUID_XMX
        value: 2g
      - name: DRUID_MAXDIRECTMEMORYSIZE
        value: 2g
      livenessProbe:
          failureThreshold: 3
          httpGet:
          path: /status/health
          port: 8083
          initialDelaySeconds: 1800
          periodSeconds: 5
      nodeConfigMountPath: /opt/druid/conf/druid/cluster/data/historical
      nodeType: historical
      podDisruptionBudgetSpec:
          maxUnavailable: 1
      readinessProbe:
          failureThreshold: 18
          httpGet:
          path: /druid/historical/v1/readiness
          port: 8083
          periodSeconds: 10
      replicas: 1
      resources:
          limits:
            cpu: 3
            memory: 6Gi
          requests:
            cpu: 1
            memory: 1Gi
      runtime.properties:
          druid.plaintextPort=8083
          druid.service=druid/historical/hot
    cold:
      druid.port: 8083
      env:
      - name: DRUID_XMS
        value: 1500m
      - name: DRUID_XMX
        value: 1500m
      - name: DRUID_MAXDIRECTMEMORYSIZE
        value: 2g
      livenessProbe:
        failureThreshold: 3
        httpGet:
          path: /status/health
          port: 8083
        initialDelaySeconds: 1800
        periodSeconds: 5
      nodeConfigMountPath: /opt/druid/conf/druid/cluster/data/historical
      nodeType: historical
      podDisruptionBudgetSpec:
        maxUnavailable: 1
      readinessProbe:
        failureThreshold: 18
        httpGet:
          path: /druid/historical/v1/readiness
          port: 8083
        periodSeconds: 10
      replicas: 1
      resources:
        limits:
          cpu: 4
          memory: 3.5Gi
        requests:
          cpu: 1
          memory: 1Gi
      runtime.properties:
        druid.plaintextPort=8083
        druid.service=druid/historical/cold
...
```

## Override default Probes
```yaml
...
  nodes:
    brokers:
      kind: Deployment
      nodeType: "broker"
      druid.port: 8088
      nodeConfigMountPath: "/opt/druid/conf/druid/cluster/query/broker"
      replicas: 2
      podDisruptionBudgetSpec:
        maxUnavailable: 1
        selector:
          matchLabels:
            app: druid
            component: broker
      livenessProbe:
        httpGet:
          path: /status/health
          port: 8088
        failureThreshold: 10
        initialDelaySeconds: 60
        periodSeconds: 30
        successThreshold: 1
        timeoutSeconds: 10
      readinessProbe:
        httpGet:
          path: /status/health
          port: 8088
        failureThreshold: 10
        initialDelaySeconds: 60
        periodSeconds: 30
        successThreshold: 1
        timeoutSeconds: 10
      resources:
        limits:
          cpu: "4"
          memory: "8Gi"
        requests:
          cpu: "2"
          memory: "4Gi"
...
```

## Configure Ingress
```yaml
...
  nodes:
    brokers:
      nodeType: "broker"
      druid.port: 8080
      ingressAnnotations:
          "nginx.ingress.kubernetes.io/rewrite-target": "/"
      ingress:
        ingressClassName: nginx # specific to ingress controllers.
        rules:
        - host: broker.myhostname.com
          http:
            paths:
            - backend:
                service:
                  name: broker_svc
                  port:
                    name: http
              path: /
              pathType: ImplementationSpecific
        tls:
        - hosts:
          - broker.myhostname.com
          secretName: tls-broker-druid-cluster
...
```

## Configure Additional Containers
```yaml
apiVersion: druid.apache.org/v1alpha1
kind: Druid
metadata:
  name: additional-containers
spec:
  additionalContainer:
    - command:
        - /bin/sh echo hello
      containerName: cluster-level
      image: hello-world
  nodes:
    brokers:
      additionalContainer:
        - command:
            - /bin/sh echo hello
          containerName: node-level
          image: hello-world
```

## Add additional configuration file into _common directory
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: hadoop-mapred-site.xml
data:
  mapred-site.xml: |
    <?xml version="1.0" encoding="UTF-8"?>
    <?xml-stylesheet type="text/xsl" href="configuration.xsl"?>
    <configuration>
        <property>
            <name>dfs.nameservices</name>
            <value>...</value>
        </property>
    </configuration>
---
apiVersion: druid.apache.org/v1alpha1
kind: Druid
metadata:
  name: druid
spec:
  extraCommonConfig:
    - name: hadoop-mapred-site.xml
      namespace: druid
...
```

## Install hadoop-dependencies With Init Container
```yaml
spec:
  volumeMounts:
    - mountPath: /opt/druid/hadoop-dependencies
      name: hadoop-dependencies
  volumes:
    - emptyDir:
        sizeLimit: 500Mi
      name: hadoop-dependencies
  additionalContainer:
    - command:
        - java
        - -cp
        - lib/*
        - -Ddruid.extensions.hadoopDependenciesDir=/hadoop-dependencies
        - org.apache.druid.cli.Main
        - tools
        - pull-deps
        - -h
        - org.apache.hadoop:hadoop-client:3.3.0
        - --no-default-hadoop
      containerName: hadoop-dependencies
      image: apache/druid:25.0.0
      runAsInit: true
      volumeMounts:
        - mountPath: /hadoop-dependencies
          name: hadoop-dependencies
```

## Secure Metadata Storage password

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: metadata-storage-password
  namespace: <NAMESPACE>
type: Opaque
data:
  METADATA_STORAGE_PASSWORD: <PASSWORD>
---
apiVersion: druid.apache.org/v1alpha1
kind: Druid
metadata:
  name: druid
spec:
  envFrom:
    - secretRef:
        name: metadata-storage-password
  nodes:
    master:
      runtime.properties: |
        # General
        druid.service=druid/coordinator

        # Metadata Storage
        druid.metadata.storage.type=<TYPE>
        druid.metadata.storage.connector.connectURI=<URI>
        druid.metadata.storage.connector.user=<USERNAME>
        druid.metadata.storage.connector.password={ "type": "environment", "variable": "METADATA_STORAGE_PASSWORD" }
...
```

## Configuration via Environment Variables (Recommended for sensitive data)

Standard Druid Docker images allow you to convert any environment variable starting with `druid_` into a configuration property (replacing `_` with `.`). This is often more reliable than the JSON replacement method found in `runtime.properties` and allows you to securely inject any configuration using Kubernetes Secrets.

**Note:** If you define a configuration here via `env`, ensure you remove it from your `runtime.properties` or `common.runtime.properties` to avoid conflicts.

### Example: Securely Injecting Secrets and Druid Properties

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: prod-druid
  namespace: druid
type: Opaque
stringData:
  # Sensitive values
  AWS_ACCESS_KEY_ID: "AKIA..."
  AWS_SECRET_ACCESS_KEY: "SECRET..."
  # You can map full property values here
  druid.metadata.storage.connector.password: "db-password"
  druid.metadata.storage.connector.connectURI: "jdbc:postgresql://..."
---
apiVersion: druid.apache.org/v1alpha1
kind: Druid
metadata:
  name: druid
spec:
  env:
    # 1. Standard Env Vars
    - name: AWS_REGION
      value: "nyc3"

    # 2. Loading Secrets
    - name: AWS_ACCESS_KEY_ID
      valueFrom:
        secretKeyRef:
          name: prod-druid
          key: AWS_ACCESS_KEY_ID
    - name: AWS_SECRET_ACCESS_KEY
      valueFrom:
        secretKeyRef:
          name: prod-druid
          key: AWS_SECRET_ACCESS_KEY

    # 3. Mapping Secrets directly to Druid Properties
    # The Docker entrypoint converts 'druid_x_y' -> 'druid.x.y'

    # Maps to: druid.metadata.storage.connector.password
    - name: druid_metadata_storage_connector_password
      valueFrom:
        secretKeyRef:
          name: prod-druid
          key: druid.metadata.storage.connector.password

    # Maps to: druid.metadata.storage.connector.connectURI
    - name: druid_metadata_storage_connector_connectURI
      valueFrom:
        secretKeyRef:
          name: prod-druid
          key: druid.metadata.storage.connector.connectURI

  nodes:
    coordinators:
      nodeType: "coordinator"
      # ...
      runtime.properties: |
        druid.service=druid/coordinator
        # Note: AWS Keys and Metadata configs are NOT listed here
        # because they are injected via the 'env' section above.
        druid.metadata.storage.type=postgresql
        druid.metadata.storage.connector.user=druid
```

### Notes

- **Environment variable expansion** (using `env` and `envFrom`) can be used for any Druid configuration property, not just passwords. This is the recommended approach for sensitive data and for properties that may change between environments.
- The JSON replacement method (`{ "type": "environment", "variable": "METADATA_STORAGE_PASSWORD" }`) is only supported for certain password fields in `runtime.properties` and is not as flexible as environment variable expansion.
- If you use both methods for the same property, the environment variable will take precedence, but it's best to avoid duplication to prevent confusion.
