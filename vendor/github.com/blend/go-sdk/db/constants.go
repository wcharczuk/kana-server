/*

Copyright (c) 2021 - Present. Blend Labs, Inc. All rights reserved
Use of this source code is governed by a MIT license that can be found in the LICENSE file.

*/

package db

import (
	"time"
)

const (
	// DefaultEngine is the default database engine.
	DefaultEngine = "pgx" // "postgres"

	// EnvVarDatabaseURL is an environment variable.
	EnvVarDatabaseURL = "DATABASE_URL"
	// EnvVarDBApplicationName is the environment variable used to set the
	// `application_name` configuration parameter in a `lib/pq` connection
	// string.
	//
	// See: https://www.postgresql.org/docs/12/runtime-config-logging.html#GUC-APPLICATION-NAME
	EnvVarDBApplicationName = "DB_APPLICATION_NAME"

	// DefaultHost is the default database hostname, typically used
	// when developing locally.
	DefaultHost = "localhost"
	// DefaultPort is the default postgres port.
	DefaultPort = "5432"
	// DefaultDatabase is the default database to connect to, we use
	// `postgres` to not pollute the template databases.
	DefaultDatabase = "postgres"

	// DefaultSchema is the default schema to connect to
	DefaultSchema = "public"

	// SSLModeDisable is an ssl mode.
	// Postgres Docs: "I don't care about security, and I don't want to pay the overhead of encryption."
	SSLModeDisable = "disable"
	// SSLModeAllow is an ssl mode.
	// Postgres Docs: "I don't care about security, but I will pay the overhead of encryption if the server insists on it."
	SSLModeAllow = "allow"
	// SSLModePrefer is an ssl mode.
	// Postgres Docs: "I don't care about encryption, but I wish to pay the overhead of encryption if the server supports it"
	SSLModePrefer = "prefer"
	// SSLModeRequire is an ssl mode.
	// Postgres Docs: "I want my data to be encrypted, and I accept the overhead. I trust that the network will make sure I always connect to the server I want."
	SSLModeRequire = "require"
	// SSLModeVerifyCA is an ssl mode.
	// Postgres Docs: "I want my data encrypted, and I accept the overhead. I want to be sure that I connect to a server that I trust."
	SSLModeVerifyCA = "verify-ca"
	// SSLModeVerifyFull is an ssl mode.
	// Postgres Docs: "I want my data encrypted, and I accept the overhead. I want to be sure that I connect to a server I trust, and that it's the one I specify."
	SSLModeVerifyFull = "verify-full"

	// DefaultConnectTimeout is the default connect timeout.
	DefaultConnectTimeout = 5 * time.Second

	// DefaultIdleConnections is the default number of idle connections.
	DefaultIdleConnections = 16
	// DefaultMaxConnections is the default maximum number of connections.
	DefaultMaxConnections = 32
	// DefaultMaxLifetime is the default maximum lifetime of driver connections.
	DefaultMaxLifetime = time.Duration(0)
	// DefaultBufferPoolSize is the default number of buffer pool entries to maintain.
	DefaultBufferPoolSize = 1024
)
