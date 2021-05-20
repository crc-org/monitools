#!/bin/sh

# Vars
MONICTL_PATH=/usr/local/bin/monictl
RESULTS_PATH="${RESULTS_PATH:-/output}"

# Check required Envs
if [[ -z "${TARGET_HOST}" ]]; then
    printf "Please set up TARGET_HOST Env\n"
    exit 1
fi

if [[ -z "${TARGET_HOST_USERNAME}" ]]; then
    printf "Please set up TARGET_HOST_USERNAME Env\n"
    exit 1
fi

if [[ -z "${TARGET_HOST_KEY_PATH}" ]]; then
    printf "Please set up TARGET_HOST_KEY_PATH Env\n"
    exit 1
fi

# Copy monictl to target host
scp -o "StrictHostKeyChecking=no" -o "UserKnownHostsFile=/dev/null" \
    -i "${TARGET_HOST_KEY_PATH}" \
    "${MONICTL_PATH}" "${TARGET_HOST_USERNAME}@${TARGET_HOST}:/home/${TARGET_HOST_USERNAME}"

# Run (one shot) monictl
#ssh -i "${TARGET_HOST_KEY_PATH}" "${TARGET_HOST_USERNAME}@${TARGET_HOST}" /home/${TARGET_HOST_USERNAME}/monictl

# Get results
#mkdir -p "${RESULTS_PATH}"
#scp -r -i "${TARGET_HOST_KEY_PATH}" \
#    "${TARGET_HOST_USERNAME}@${TARGET_HOST}:/home/${TARGET_HOST_USERNAME}/monictl-results" "${RESULTS_PATH}"