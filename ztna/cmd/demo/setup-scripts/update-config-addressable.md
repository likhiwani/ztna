# Purpose

This script updates the echo service hosting to allow addressing the individual servers.

# Prerequisites

You need at least one controller and an edge router running. for this to work.
You can use the quick-start script found [here](https://ztna-core/ztna/tree/release-next/quickstart).

# Setup

## Ensure we're logged into the controller.

```action:ziti-login allowRetry=true
ziti edge login
```

<!--action:keep-session-alive interval=1m quiet=false-->

## Update the echo-host config

```action:ziti-update-config name=echo-host
{
    "terminators" : [
        {
            "address" : "localhost",
            "port" : 1234,
            "protocol" : "tcp",
            "portChecks" : [
                {
                     "address" : "localhost:2234",
                     "interval" : "1s",
                     "timeout" : "100ms",
                     "actions" : [
                         { "trigger" : "fail", "action" : "mark unhealthy" },
                         { "trigger" : "pass", "action" : "mark healthy" }
                     ]
                }
           ],
           "listenOptions" : {
                "identity" : "echo-host-1"
           }
        },
        {
            "address" : "localhost",
            "port" : 1235,
            "protocol" : "tcp",
            "portChecks" : [
                {
                     "address" : "localhost:2235",
                     "interval" : "1s",
                     "timeout" : "100ms",
                     "actions" : [
                         { "trigger" : "fail", "action" : "mark unhealthy" },
                         { "trigger" : "pass", "action" : "mark healthy" }
                     ]
                }
           ],
           "listenOptions" : {
                "identity" : "echo-host-2"
           }
        }
    ]
}
```

## Summary

You should now be to run the echo server with

```
ziti demo echo-server -p 1234
```

and the zcat client using

```
ziti demo zcat -i zcat.json ziti:echo-host-1@echo
```
