#!/usr/bin/env sh

if [ -z "${DATA_DIR}" -o -z "${DATA_FILE}" ]; then
  die "DATA_DIR and DATA_FILE environment variables must be set"
fi

# initialize data dir if not exists already
if [ ! -e "${DATA_DIR}/${DATA_FILE}" ]; then
  echo "Initializing ${DATA_DIR}/${DATA_FILE}"
  ./tigerbeetle format --cluster=0 --replica=0 "${DATA_DIR}/${DATA_FILE}"
fi

PORT=${PORT:-3000}

./tigerbeetle start --addresses=0.0.0.0:${PORT} ${DATA_DIR}/${DATA_FILE}
