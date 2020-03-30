module mrfox-cron-worker

go 1.14

require (
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738
	go.mongodb.org/mongo-driver v1.3.1
	mrfox-cron-common v0.0.0
	mrfox-cron-facade v0.0.0
)

replace (
	mrfox-cron-common v0.0.0 => ../common
	mrfox-cron-facade v0.0.0 => ../facade
)
