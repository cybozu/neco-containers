#! /bin/sh

echo "Run placemat"
echo "Nothing to do right now, so skip"

echo "Run entrypoint"
./entrypoint.sh

if [ -f /home/runner/delete-immediately ]; then
    echo "Label runners with current time"
    kubectl label runners.actions.summerwind.dev ${RUNNER_NAME} delete-at=$(date +%Y%m%d%H%M%S)
else
    echo "Label runners with current time + 1m"
    kubectl label runners.actions.summerwind.dev ${RUNNER_NAME} delete-at=$(date -d "1 minutes" +%Y%m%d%H%M%S)
fi

echo "Wait until delete-at"
while true
do
    DELETE_AT=$(kubectl get runners.actions.summerwind.dev ${RUNNER_NAME} -o jsonpath='{.metadata.labels.delete-at}')
    NOW=$(date +%Y%m%d%H%M%S)
    if [ -n "${DELETE_AT}" ] && [ ${NOW} -gt ${DELETE_AT} ]; then
        echo "Delete ${RUNNER_NAME}"
        kubectl delete runners.actions.summerwind.dev ${RUNNER_NAME}
    fi
    echo "sleeping..."
    sleep 30
done
