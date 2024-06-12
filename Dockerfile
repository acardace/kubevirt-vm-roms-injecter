FROM quay.io/kubevirt/sidecar-shim:v1.2.2

COPY ./onDefineDomain /usr/bin

ENTRYPOINT ["/sidecar-shim", "--version", "v1alpha2"]
