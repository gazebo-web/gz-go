<div align="center">
  <img src="./assets/logo.png" width="200" alt="Ignition Robotics" />
  <h1>Ignition Robotics</h1>
  <p>Ignition GO is a general purpose golang library that encapsulates a set of common functionality for a webserver.</p>
</div>

## Getting started

## Installing

## Contribute

### Environment variables
- IGN_SSL_CERT : Path to an SSL certificate file. This is used for local SSL testing and development.
- IGN_SSL_KEY : Path to an SSL key. THis is used for local SSL testing and development
- IGN_DB_USERNAME : Username for the database connection.
- IGN_DB_PASSWORD : Password for the database connection.
- IGN_DB_ADDRESS : URL address for the database server.
- IGN_DB_NAME : Name of the database to use on the database sever.
- IGN_DB_LOG : Controls whether or not database transactions generate log output. Set to true to enable database logging. This environment variable is optional, and database logging will default to off expect for tests.
- IGN_DB_MAX_OPEN_CONNS : Max number of open connections in connections pool. A value <= 0 means unlimited connections. Tip: You can learn max_connections of your mysql by running this query: SHOW VARIABLES LIKE "max_connections";
- IGN_GA_TRACKING_ID : Google Analytics Tracking ID to use. If not set, then GA will not be enabled. The format is UA-XXXX-Y.
- IGN_GA_APP_NAME : Google Analytics Application Name. If not set, then GA will not be enabled.
- IGN_GA_CAT_PREFIX : (optional) A string to use as a prefix to Google Analytics Event Category.
- IGN_ROLLBAR_TOKEN: (optional) Rollbar authentication token. If valid, then log messages will be sent to rollbar.com. It is recommended NOT to use rollbar during local development.
- IGN_ROLLBAR_ENV: (optional) Rollbar environment string, such as "staging" or "production".
- IGN_ROLLBAR_ROOT: (optional) Path to the application code root, not including the final slash. Such as bitbucket.org/ignitionrobotics/ign-fuelserver
- IGN_LOGGER_LOG_STDOUT: (optional) Controls whether or not logs will be also sent to stdout/err. If missing, a false value will be used.

## Documentation

## Building

## Roadmap
