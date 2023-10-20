# dependencies

go get github.com/envoyproxy/go-control-plane

https://grpc.io/docs/languages/go/quickstart/
https://gokhan-karadas1992.medium.com/envoy-remote-access-log-collector-84e4cde6375c
https://istio.io/latest/docs/tasks/observability/logs/access-log/
https://tetrate.io/blog/observe-service-mesh-with-skywalking-and-envoy-access-log-service/
https://istio.io/latest/docs/tasks/observability/logs/access-log/

https://www.envoyproxy.io/docs/envoy/latest/api-v3/data/accesslog/v3/accesslog.proto#envoy-v3-api-msg-data-accesslog-v3-httpaccesslogentry

https://www.cncf.io/blog/2019/10/15/extend-kubernetes-via-a-shared-informer/

# envoy configuration

curl -L https://istio.io/downloadIstio | sh -
export PATH="$PATH:/data1/home/fidok/go_workspace/service-mesh-api-discovery/istio-1.19.3/bin"
istioctl install
kubectl label namespace default istio-injection=enabled
istioctl manifest install --set meshConfig.enableEnvoyAccessLogService=true --set meshConfig.defaultConfig.envoyAccessLogService.address=envoy-receiver-svc.default.svc.cluster.local:65000 --set meshConfig.accessLogFile=/dev/stdout
kubectl -n istio-system get cm istio -o yaml

kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.19/samples/bookinfo/platform/kube/bookinfo.yaml

eval $(minikube docker-env)

docker build -f build/Dockerfile . -t receiver:latest

data:
mesh: |-
accessLogEncoding: JSON
accessLogFile: ""
accessLogFormat: ""
connectTimeout: 15s
enableEnvoyAccessLogService: true
defaultConfig:
envoyAccessLogService:
address: envoy-accesslog-collector.platform:9001
