# Build the manager binary
FROM registry.access.redhat.com/ubi8/go-toolset:1.14.12 as builder

USER root
WORKDIR /workspace
COPY . .
RUN make build

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL MAINTAINER "CodeReady Containers <devtools-cdk@redhat.com>"

COPY --from=builder /workspace/bin/linux/monictl /usr/local/bin/
COPY --from=builder /workspace/images/build/entrypoint.sh /usr/local/bin/

RUN microdnf install -y openssh-clients

ENTRYPOINT ["entrypoint.sh"]
