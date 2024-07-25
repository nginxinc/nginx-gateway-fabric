#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source scripts/vars.env

SCRIPT=run-tests.sh
if [ "${NFR}" = "true" ]; then
    SCRIPT=run-nfr-tests.sh
fi

gcloud compute scp --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} ${SCRIPT_DIR}/vars.env username@${RESOURCE_NAME}:~

gcloud compute ssh --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} username@${RESOURCE_NAME} \
    --command="export START_LONGEVITY=${START_LONGEVITY} &&\
        export STOP_LONGEVITY=${STOP_LONGEVITY} &&\
		source /etc/profile &&\
        bash -s" < ${SCRIPT_DIR}/remote-scripts/${SCRIPT}; retcode=$?

if [ $retcode -ne 0 ]; then
	echo "Error running tests on VM"
	exit 1
fi

if [ "${NFR}" = "true" ]; then
    ## Use rsync if running locally (faster); otherwise if in the pipeline don't download an SSH config
    if [ "${CI}" = "false" ]; then
      gcloud compute config-ssh --ssh-config-file ngf-gcp.ssh > /dev/null
      rsync -ave 'ssh -F ngf-gcp.ssh' username@${RESOURCE_NAME}.${GKE_CLUSTER_ZONE}.${GKE_PROJECT}:~/nginx-gateway-fabric/tests/results .
    else
      gcloud compute scp --zone ${GKE_CLUSTER_ZONE} --project=${GKE_PROJECT} --recurse username@${RESOURCE_NAME}:~/nginx-gateway-fabric/tests/results .
    fi
fi

## If tearing down the longevity test, we need to collect logs from gcloud and add to the results
if [ "${STOP_LONGEVITY}" = "true" ]; then
    version=${NGF_VERSION}
    if [ "$version" = "" ]; then
        version=${TAG}
    fi

    results="${SCRIPT_DIR}/../results/longevity/$version/$version.md"
    printf "\n## Error Logs\n\n" >> $results

    ## ngf error logs
    ngfErrText=$(gcloud logging read --project=${GKE_PROJECT} 'resource.labels.cluster_name='"${RESOURCE_NAME}"' AND resource.type=k8s_container AND resource.labels.container_name=nginx-gateway AND labels."k8s-pod/app_kubernetes_io/instance"=ngf-longevity AND severity=ERROR AND SEARCH("error")' --format "value(textPayload)")
    ngfErrJSON=$(gcloud logging read --project=${GKE_PROJECT} 'resource.labels.cluster_name='"${RESOURCE_NAME}"' AND resource.type=k8s_container AND resource.labels.container_name=nginx-gateway AND labels."k8s-pod/app_kubernetes_io/instance"=ngf-longevity AND severity=ERROR AND SEARCH("error")' --format "value(jsonPayload)")
    printf "### nginx-gateway\n$ngfErrText\n$ngfErrJSON\n\n" >> $results

    ## nginx error logs
    ngxErr=$(gcloud logging read --project=${GKE_PROJECT} 'resource.labels.cluster_name='"${RESOURCE_NAME}"' AND resource.type=k8s_container AND resource.labels.container_name=nginx AND labels."k8s-pod/app_kubernetes_io/instance"=ngf-longevity AND severity=ERROR AND SEARCH("`[warn]`") OR SEARCH("`[error]`") OR SEARCH("`[emerg]`")' --format "value(textPayload)")
    printf "### nginx\n$ngxErr\n\n" >> $results

    ## nginx non-200 responses (also filter out 499 since wrk cancels connections)
    ngxNon200=$(gcloud logging read --project=${GKE_PROJECT} 'resource.labels.cluster_name='"${RESOURCE_NAME}"' AND resource.type=k8s_container AND resource.labels.container_name=nginx AND labels."k8s-pod/app_kubernetes_io/instance"=ngf-longevity AND "GET" "HTTP/1.1" -"200" -"499" -"client prematurely closed connection"' --format "value(textPayload)")
    printf "$ngxNon200\n\n" >> $results
fi
