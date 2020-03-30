module mrfox-cron-master

go 1.14

require (
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	go.mongodb.org/mongo-driver v1.3.1
	gopkg.in/yaml.v2 v2.2.4 // indirect
	mrfox-cron-common v0.0.0
	mrfox-cron-facade v0.0.0
)

replace (
	mrfox-cron-common v0.0.0 => ../common
	mrfox-cron-facade v0.0.0 => ../facade
)
