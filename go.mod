module github.com/layer5io/meshery-octarine

require (
	cloud.google.com/go v0.39.0 // indirect
	github.com/DataDog/zstd v1.4.0 // indirect
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/Shopify/sarama v1.22.1 // indirect
	github.com/apache/thrift v0.12.0 // indirect
	github.com/dgryski/go-sip13 v0.0.0-20190329191031-25c5027a8c7b // indirect
	github.com/envoyproxy/go-control-plane v0.8.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/googleapis v1.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/golang/mock v1.3.1 // indirect
	github.com/golang/protobuf v1.3.5
	github.com/google/btree v1.0.0 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/google/pprof v0.0.0-20190515194954-54271f7e092f // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gorilla/mux v1.7.2 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/kisielk/errcheck v1.2.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pty v1.1.4 // indirect
	github.com/lyft/protoc-gen-validate v0.0.14 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/openzipkin/zipkin-go v0.1.6 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3 // indirect
	github.com/prometheus/common v0.4.1 // indirect
	github.com/prometheus/procfs v0.0.2 // indirect
	github.com/prometheus/tsdb v0.8.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	go.opencensus.io v0.22.0 // indirect
	golang.org/x/crypto v0.0.0-20190530122614-20be4c3c3ed5 // indirect
	golang.org/x/exp v0.0.0-20190510132918-efd6b22b2522 // indirect
	golang.org/x/image v0.0.0-20190523035834-f03afa92d3ff // indirect
	golang.org/x/lint v0.0.0-20190409202823-959b441ac422 // indirect
	golang.org/x/mobile v0.0.0-20190509164839-32b2708ab171 // indirect
	golang.org/x/mod v0.1.0 // indirect
	golang.org/x/net v0.0.0-20190603091049-60506f45cf65
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sys v0.0.0-20190602015325-4c4f7f33c9ed // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	golang.org/x/tools v0.0.0-20190603231351-8aaa1484dc10 // indirect
	google.golang.org/appengine v1.6.0 // indirect
	google.golang.org/genproto v0.0.0-20190530194941-fb225487d101 // indirect
	google.golang.org/grpc v1.21.1
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	honnef.co/go/tools v0.0.0-20190604153307-63e9ff576adb // indirect
	k8s.io/api v0.0.0-20190602205700-9b8cae951d65 // indirect
	k8s.io/apimachinery v0.0.0-20190602183612-63a6072eb563
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/klog v0.3.2 // indirect
	k8s.io/utils v0.0.0-20190529001817-6999998975a7 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace github.com/golang/sync => golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4

replace github.com/golang/lint v0.0.0-20190409202823-959b441ac422 => golang.org/x/lint v0.0.0-20190409202823-959b441ac422

replace sourcegraph.com/sourcegraph/go-diff v0.5.1 => github.com/sourcegraph/go-diff v0.5.1

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d

go 1.13
