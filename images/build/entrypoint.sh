#!/bin/sh

# Vars
MONICTL_PATH=/usr/local/bin/monictl
RESULTS_PATH="${RESULTS_PATH:-/output}"
MONICTL_REPETITIONS="${MONICTL_REPETITIONS:-5}"
MONICTL_INTERVAL="${MONICTL_INTERVAL:-1}"

if [ "${DEBUG:-}" = "true" ]; then
    set -xuo 
fi

# Validate conf
validate=true
[[ -z "${TARGET_HOST}" ]] \
    && echo "TARGET_HOST requried" \
    && validate=false

[[ -z "${TARGET_HOST_USERNAME}" ]] \
    && echo "TARGET_HOST_USERNAME requried" \
    && validate=false

[[ -z "${TARGET_HOST_KEY_PATH}" && -z "${TARGET_HOST_PASSWORD}" ]] \
    && echo "TARGET_HOST_KEY_PATH or TARGET_HOST_PASSWORD required" \
    && validate=false
[[ $validate == false ]] && exit 1

# Set SCP / SSH command
NO_STRICT='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null'
[[ ! -z "${TARGET_HOST_KEY_PATH}" ]] \
    && SCP="scp ${NO_STRICT} -i ${TARGET_HOST_KEY_PATH}" \
        || SCP="sshpass -p ${TARGET_HOST_PASSWORD} scp ${NO_STRICT}" \
    && SSH="ssh ${NO_STRICT} -i ${TARGET_HOST_KEY_PATH}" \
        || SSH="sshpass -p ${TARGET_HOST_PASSWORD} ssh ${NO_STRICT}"

# Create execution folder 
EXECUTION_FOLDER="/home/${TARGET_HOST_USERNAME}/monitools/${RANDOM}"
DATA_FOLDER="${EXECUTION_FOLDER}/data"
$SSH "${TARGET_HOST_USERNAME}@${TARGET_HOST}" "mkdir -p ${EXECUTION_FOLDER} && mkdir -p ${DATA_FOLDER}" 

# Copy monictl to target host
$SCP "${MONICTL_PATH}" "${TARGET_HOST_USERNAME}@${TARGET_HOST}:${EXECUTION_FOLDER}"

# Run (one shot) monictl
MONITOOL_EXEC="${EXECUTION_FOLDER}/monictl -d ${DATA_FOLDER} -n ${MONICTL_REPETITIONS} -s ${MONICTL_INTERVAL}"
$SSH "${TARGET_HOST_USERNAME}@${TARGET_HOST}" "${MONITOOL_EXEC}"

# Get results
mkdir -p "${RESULTS_PATH}"
$SCP -r "${TARGET_HOST_USERNAME}@${TARGET_HOST}:${DATA_FOLDER}" "${RESULTS_PATH}"

# Remove remote execution fodler
$SSH "${TARGET_HOST_USERNAME}@${TARGET_HOST}" rm -rf "${EXECUTION_FOLDER}"
