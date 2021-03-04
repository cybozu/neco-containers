#! /bin/sh

echo "Run placemat"
echo "Nothing to do right now, so skip"

echo "Register myself as self-hosted runner"
export RUNNER_TOKEN=$(register-actions)

echo "Run entrypoint"
./entrypoint.sh

if [ -f /tmp/failed ]; then
    echo "Label pods with current time + 1m"
    kubectl label pod ${POD_NAME} delete-at=$(date -d "1 minutes" +%Y%m%d%H%M%S)
else
    echo "Label pods with current time"
    kubectl label pod ${POD_NAME} delete-at=$(date +%Y%m%d%H%M%S)
fi

echo "Wait until delete-at"
while true
do
    DELETE_AT=$(kubectl get pod ${POD_NAME} -o jsonpath='{.metadata.labels.delete-at}')
    NOW=$(date +%Y%m%d%H%M%S)
    if [ -n "${DELETE_AT}" ] && [ ${NOW} -gt ${DELETE_AT} ]; then
        echo "Delete ${POD_NAME}"
        kubectl delete pod ${POD_NAME}
    fi
    echo "sleeping..."
    sleep 30
done
