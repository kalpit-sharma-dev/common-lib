module gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6

go 1.15

replace (
	github.com/afex/hystrix-go => github.com/ContinuumLLC/hystrix-go v0.0.0-20190403132145-d82962fc32a8
	github.com/samuel/go-zookeeper => github.com/ContinuumLLC/go-zookeeper v1.0.0
)

require (
	cloud.google.com/go v0.84.0 // indirect
	github.com/Comcast/go-leaderelection v0.0.0-20181102191523-272fd9e2bddc
	github.com/DATA-DOG/go-sqlmock v1.4.1
	github.com/DataDog/zstd v1.4.9-0.20210607132535-4fa4b6b2bd43 // indirect
	github.com/OneOfOne/xxhash v1.2.8
	github.com/Shopify/sarama v1.19.1-0.20181205181954-9daa115cef80
	github.com/StackExchange/wmi v0.0.0-20181212234831-e0a55b97c705
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/alicebob/miniredis v2.5.0+incompatible
	github.com/avast/retry-go v2.6.0+incompatible
	github.com/aws/aws-sdk-go-v2 v1.16.2
	github.com/aws/aws-sdk-go-v2/config v1.15.3
	github.com/aws/aws-sdk-go-v2/service/ses v1.14.0
	github.com/aws/aws-xray-sdk-go v1.6.0
	github.com/aws/smithy-go v1.11.2
	github.com/bsm/sarama-cluster v2.1.16-0.20181008124012-8cd6c692710b+incompatible
	github.com/cavaliercoder/grab v2.0.1-0.20190724181540-228f991ef22e+incompatible
	github.com/cheekybits/genny v1.0.0
	github.com/confluentinc/confluent-kafka-go v1.8.2
	github.com/coocood/freecache v0.0.0-20170527025705-a47e26eb67ac
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20190423183735-731ef375ac02
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/eapache/go-resiliency v1.1.0
	github.com/eapache/queue v1.1.1-0.20180227141424-093482f3f8ce // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/gammazero/deque v0.1.0 // indirect
	github.com/gammazero/workerpool v1.0.0
	github.com/go-ole/go-ole v1.2.2-0.20181122093336-ae2e2a20879a // indirect
	github.com/go-playground/validator/v10 v10.9.0
	github.com/go-redis/redis v0.0.0-20190503082931-75795aa4236d
	github.com/gocql/gocql v0.0.0-20211015133455-b225f9b53fa1
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/gomodule/redigo v1.8.5 // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.7.4
	github.com/jarcoal/httpmock v1.0.8
	github.com/jinzhu/copier v0.3.2
	github.com/jinzhu/gorm v1.9.8
	github.com/jmoiron/sqlx v1.3.0
	github.com/kardianos/service v1.0.1-0.20190326161025-0e5bec1b9eec
	github.com/kennygrant/sanitize v1.2.4
	github.com/lib/pq v1.10.0
	github.com/maraino/go-mock v0.0.0-20180321183845-4c74c434cd3a
	github.com/mattn/go-ieproxy v0.0.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.6.0
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/robfig/cron v1.0.1-0.20170526150127-736158dc09e1
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/scylladb/gocqlx v0.0.0-20180515120735-5526e6046474
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/snowflakedb/gosnowflake v1.1.7-0.20180403151706-9baa3151d076
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/yuin/gopher-lua v0.0.0-20210529063254-f4c35e4016d9 // indirect
	gitlab.kksharmadevdev.com/platform/platform-api-model v0.0.0-20220311122951-55cf5d733d37
	go.uber.org/atomic v1.9.0
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/net v0.0.0-20211209124913-491a49abca63
	golang.org/x/sys v0.0.0-20210806184541-e5e7981a1069
	google.golang.org/genproto v0.0.0-20210726143408-b02e89920bf0 // indirect
	google.golang.org/protobuf v1.27.1
	gopkg.in/ini.v1 v1.39.3
	gopkg.in/natefinch/lumberjack.v2 v2.0.0-20170531160350-a96e63847dc3
	gopkg.in/urfave/cli.v1 v1.20.0
)
