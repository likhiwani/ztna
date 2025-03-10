# Purpose

This script sets up hosting for an echo service which is hosted by a router-embedded tunneler.

# Prerequisites

You need at least one controller and an edge router running. for this to work.
You can use the quick-start script found [here](https://ztna-core/ztna/tree/release-next/quickstart).

# Setup

## Ensure we're logged into the controller.

```action:ziti-login allowRetry=true
ziti edge login
```

<!--action:keep-session-alive interval=1m quiet=false-->

## Remove any entities from previous runs.

```action:ziti
ziti edge delete service echo
ziti edge delete config echo-host
ziti edge delete identities echo-host-1 echo-host-2
ziti edge delete service-policies echo-bind
ziti edge delete edge-router-policies echo
ziti edge delete service-edge-router-policies echo 
```

## Create the echo-host config

```action:ziti-create-config name=echo-host type=host.v2
{
    "terminators" : [
        {
            "address" : "localhost",
            "port" : 1234,
            "protocol" : "tcp"   
        }
    ]
}
```

## Create the echo service

```action:ziti
ziti edge create service echo -c echo-host -a echo
```

## Update edge-routers

Make sure demo edge routers are tunneler enabled and the associated identity has the `echo-host` attribute.
Only routers with the `demo` role attribute will be updated.

```action:ziti-for-each type=edge-routers minCount=1 maxCount=2 filter='anyOf(roleAttributes)="demo"'
ziti edge update edge-router ${entityName} --tunneler-enabled
ziti edge update identity ${entityName} --role-attributes echo-host 
```

## Configure policies

```action:ziti
ziti edge create service-policy echo-bind Bind --service-roles @echo --identity-roles #echo-host
ziti edge create edge-router-policy echo --identity-roles #echo --edge-router-roles #all
ziti edge create service-edge-router-policy echo --service-roles @echo --edge-router-roles #all
```

## Summary

You should now be to run the echo server with

```
ziti demo echo-server -p 1234
```

and the zcat client using

```
ziti demo zcat -i zcat.json ziti:echo
```
