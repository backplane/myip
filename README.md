# myip

This is an HTTP endpoint which returns the requester's own IP address in JSON format.

## Config

In addition to the CLI flags described below, the following envvars can configure the endpoint:

| ENVVAR            | Default        | Description                                                                                |
| ----------------- | -------------- | ------------------------------------------------------------------------------------------ |
| `LOG_LEVEL`       | `INFO`         | how verbosely to log, one of: `DEBUG`, `INFO`, `WARN`, `ERROR`                             |
| `LISTEN_ADDR`     | `0.0.0.0:8000` | the host and port number to receive request on                                             |
| `TRUST_XFF`       | `false`        | trust X-Forwarded-For headers in the request (only enable if running behind a proxy)       |
| `TRUSTED_PROXIES` | ``             | comma-separated list of IP blocks (in CIDR-notation) that upstream proxy request come from |

## Usage

This is the output of the program when it is invoked with the `-h` argument:

````
NAME:
   myip - HTTP endpoint that reports the user's IP address back to the user

USAGE:
   myip [global options]

VERSION:
   dev

GLOBAL OPTIONS:
   --loglevel string        how verbosely to log, one of: DEBUG, INFO, WARN, ERROR (default: "INFO") [$LOG_LEVEL]
   --listenaddr string      IP address and port to listen on (default: "0.0.0.0:8000") [$LISTEN_ADDR]
   --trustxff               trust X-Forwarded-For headers in the request (only enable if running behind a proxy) (default: false) [$TRUST_XFF]
   --trustedproxies string  comma-separated list of IP blocks (in CIDR-notation) that upstream proxy request come from [$TRUSTED_PROXIES]
   --help, -h               show help
   --version, -v            print the version```
````
