# Loki Operator Development Guide

This file provides detailed architectural guidance and development context for AI agents working with the Loki Operator codebase.

## Overview

The Loki Operator is a Kubernetes controller that manages Loki deployments using Custom Resource Definitions (CRDs).

## Directory Structure

**`operator/`**: the operator sub-project

- **`api/`**: Custom Resource Definitions and API types
  - `loki/v1/`: Stable API version with core resource types
    - `lokistack_types.go`: Defines LokiStack custom resource for complete Loki deployments
    - `alertingrule_types.go`: Manages LogQL-based alerting rules
    - `recordingrule_types.go`: Defines recording rules for pre-computed metrics
    - `rulerconfig_types.go`: Configures the ruler component behavior
  - `loki/v1beta1/`: Beta API version for experimental features

- **`cmd/`**: Operator executables and utilities
  - `loki-operator/`: Main operator controller binary
  - `loki-broker/`: CLI tool for operator management and debugging
  - `size-calculator/`: Storage size calculation utility for capacity planning

- **`internal/`**: Core operator implementation (not part of public API)
  - `controller/`: Kubernetes controller reconciliation logic
  - `config/`: Configuration management and validation
  - `manifests/`: Kubernetes manifest generation and templating
  - `operator/`: Core operator business logic and resource management
  - `validation/`: Resource validation and admission control
  - `sizes/`: Storage sizing algorithms and calculations

- **`config/`**: Kubernetes deployment configurations
  - `crd/`: Custom Resource Definition bases
  - `rbac/`: Role-Based Access Control configurations
  - `manager/`: Operator deployment manifests
  - `samples/`: Example Custom Resource configurations
  - `kustomize/`: Kustomize overlays for different environments

- **`bundle/`**: Operator Lifecycle Manager (OLM) packaging
  - Supports multiple deployment variants:
    - `community`: Standard Grafana distribution
    - `community-openshift`: OpenShift-compatible community version
    - `openshift`: Red Hat certified OpenShift distribution
  - Contains ClusterServiceVersion, package manifests, and bundle metadata

## Key Features

- **Multi-tenant Support**: Isolates log streams by tenant with configurable authentication
- **Flexible Storage**: Supports object storage (S3, GCS, Azure), local storage, and hybrid configurations
- **Auto-scaling**: Horizontal Pod Autoscaler integration for dynamic scaling
- **Security**: Integration with OpenShift authentication, RBAC, and network policies
- **Monitoring**: Built-in Prometheus metrics and Grafana dashboard integration
- **Gateway Component**: Optional log routing and tenant isolation layer

## Deployment Variants

1. **Community** (`VARIANT=community`):
   - Registry: `docker.io/grafana`
   - Standard Kubernetes deployment
   - Flexible configuration options
   - Community support channels

2. **Community-OpenShift** (`VARIANT=community-openshift`):
   - Optimized for OpenShift but community-supported
   - Enhanced security contexts
   - OpenShift-specific networking configurations

3. **OpenShift** (`VARIANT=openshift`):
   - Registry: `quay.io/openshift-logging`
   - Namespace: `openshift-operators-redhat`
   - Full Red Hat support and certification
   - Tight integration with OpenShift Logging stack

## Build and Development Commands


### Development

``` bash
make cli                       # Build loki-broker CLI binary
make manager                   # Build manager binary
make size-calculator           # Build size-calculator binary
make go-generate               # Run go generate
make generate                  # Generate controller and crd code
make manifests                 # Generate manifests e.g. CRD, RBAC etc.
make test                      # Run tests
make test-unit-prometheus      # Run prometheus unit tests
make scorecard                 # Run scorecard tests for all bundles (community, community-openshift, openshift)
make lint                      # Run golangci-lint on source code.
make lint-fix                  # Attempt to automatically fix lint issues in source code.
make lint-prometheus           # Run promtool check against recording rules and alerts.
make fmt                       # Format the source code.
make oci-build                 # Build the image
make oci-push                  # Push the image
make bundle-all                # Generate both bundles.
make bundle                    # Generate variant bundle manifests and metadata, then validate generated files.
make bundle-build              # Build the community bundle image

### Deployment
make quickstart                # Quickstart full dev environment on local kind cluster
make quickstart-cleanup        # Cleanup for quickstart set up
make run                       # Run against the configured Kubernetes cluster in ~/.kube/config
make install                   # Install CRDs into a cluster
make uninstall                 # Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
make deploy                    # Deploy controller in the configured Kubernetes cluster in ~/.kube/config
make undeploy                  # Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
make olm-deploy                # Deploy the operator bundle and the operator via OLM into an Kubernetes cluster selected via KUBECONFIG.
make olm-upgrade               # Upgrade the operator bundle and the operator via OLM into an Kubernetes cluster selected via KUBECONFIG.
make olm-undeploy              # Cleanup deployments of the operator bundle and the operator via OLM on an OpenShift cluster selected via KUBECONFIG.
make deploy-size-calculator    # Deploy storage size calculator (OpenShift only!)
make deploy-size-calculator    # Deploy storage size calculator (OpenShift only!)
make undeploy-size-calculator  # Undeploy storage size calculator
make oci-build-calculator      # Build the calculator image
make oci-push-calculator       # Push the calculator image

### Website
make web-pre                   # Build the markdown API files of the loki-operator.dev website
make web                       # Run production build of the loki-operator.dev website
make web-serve                 # Run local preview version of the loki-operator.dev website
```

## Contributing to Operator Code

1. **Development Environment Setup**:
   ```bash
   # Prerequisites: Go 1.21+, Docker/Podman, kind or OpenShift cluster
   git clone https://github.com/grafana/loki.git
   cd loki/operator
   make quickstart                # Sets up local environment
   ```

2. **Development Workflow**:
   ```bash
   # Make changes to API types, controllers, or manifests
   make test                     # Run unit tests
   ```

3. **Adding New API Fields**:
   - Modify types in `api/loki/v1/*.go`
   - Run `make generate manifests` to update generated code
   - Add validation logic in `internal/validation/`
   - Update controller reconciliation in `internal/controller/`
   - Write comprehensive unit tests

4. **Adding New Features**:
   - Extend controller logic in `internal/controller/lokistack/`
   - Add manifest generation in `internal/manifests/`
   - Update configuration handling in `internal/config/`
   - Add feature flags if needed
   - Document in operator documentation

5. **Running the operator locally**:
   ```bash
   # Local testing workflow
   make install                  # Install CRDs
   make run                      # Run operator locally
   # Apply sample CustomResources in another terminal
   kubectl apply -f config/samples/
   ```

6. **Bundle and Release Process**:
   ```bash
   # Test bundle generation for all variants
   make bundle VARIANT=community
   make bundle VARIANT=community-openshift
   make bundle VARIANT=openshift
   ```

## Common Development Tasks

- **Adding New CRD Field**: Modify `*_types.go`, run `make generate manifests`
- **Updating Controller Logic**: Edit `internal/controller/`, ensure proper reconciliation
- **Adding Storage Backend**: Extend `internal/manifests/storage.go`
- **Enhancing Validation**: Update `internal/validation/` with new rules
- **Supporting New Loki Version**: Update manifests and test compatibility

## Troubleshooting Development Issues

```bash
# Debug operator logs
kubectl logs -f deployment/loki-operator-controller-manager -n loki-operator-system

# Check CRD status
kubectl describe lokistack my-stack -n my-namespace

# Validate generated manifests
kubectl apply --dry-run=client -f config/samples/

# Test bundle locally
operator-sdk run bundle-upgrade docker.io/grafana/loki-operator-bundle:latest
```
