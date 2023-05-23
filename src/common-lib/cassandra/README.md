<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Cassandra - [WIP]

Wrapper around gocql library for accessing Apache Cassandra database.

**Import Statement**

```go
import "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra"
```

# Connecting with AWS Keyspaces

To establish connection with AWS Keyspaces a few fields of a default DbConfig instance require modifications as specified below -

**TLS certificate**  
Transport Layer Security (TLS) certificate is required to establish secure connections with clients

```go
sslOptions:=&gocql.SslOptions{
    CaPath: caPath, // Path to TLS-certificate file
},
```

**Configure AWS authentication information**  

```go
 import "github.com/aws/aws-sigv4-auth-cassandra-gocql-driver-plugin/sigv4"
```

```go
//initializes authenticator with aws authentication information
var auth sigv4.AwsAuthenticator = sigv4.NewAwsAuthenticator()
auth.Region = "awsRegion"
auth.AccessKeyId = "AccessKeyID"
auth.SecretAccessKey = "SecretAccessKey"
auth.SessionToken = "SessionToken"
```

**Configuring DBConfig instance for AWS Keyspaces**  

```go
    // An instance of DbConfig with example configurations for AWS Keyspaces
    CassandraConfig:= cassandra.DbConfig{
        // add the Amazon Keyspaces service endpoint
        Hosts:                    ["cassandra.us-east-1.amazonaws.com:9142",]
        Keyspace:                 "keyspace-name",
        Authenticator:            auth,
        SslOpts:                  &gocql.SslOptions{
        CaPath:                 caPath,
        },
        // Override default Consistency to LocalQuorum
        Consistency:gocql.LocalQuorum,
        // ConnectTimeout may differ due to latency
        ConnectTimeout = 600 * time.Millisecond 
        DisableInitialHostLookup: false,
}

```
