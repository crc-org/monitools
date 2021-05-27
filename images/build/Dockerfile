# Build the manager binary
FROM registry.access.redhat.com/ubi8/go-toolset:1.14.12 as builder

USER root
WORKDIR /workspace
COPY . .
RUN make build

FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL MAINTAINER "CodeReady Containers <devtools-cdk@redhat.com>"

COPY --from=builder /workspace/bin/linux/monictl /workspace/images/build/entrypoint.sh /usr/local/bin/

ENV EPEL https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm

RUN rpm -ivh ${EPEL} \ 
    && microdnf --enablerepo=epel install -y openssh-clients sshpass

ENTRYPOINT ["entrypoint.sh"]
