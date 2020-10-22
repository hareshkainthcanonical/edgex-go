#!/bin/sh -e

# check if a value has been set for edgex-startup-duration

# NOTE: need to implement regex to check if value for both variables only contain a number

EDGEX_STARTUP_DURATION_VAL=$(snapctl get app-service-config-startup-duration)
if [ -n "$EDGEX_STARTUP_DURATION_VAL" ]; then
    # then export the env var as false to turn everything off
    EDGEX_STARTUP_DURATION=EDGEX_STARTUP_DURATION_VAL
    export EDGEX_STARTUP_DURATION
fi

# check if a value has been set for edgex-startup-interval

EDGEX_STARTUP_INTERVAL_VAL=$(snapctl get app-service-config-startup-interval)
if [ -n "$EDGEX_STARTUP_INTERVAL_VAL" ]; then
    # then export the env var as false to turn everything off
    EDGEX_STARTUP_INTERVAL=EDGEX_STARTUP_INTERVAL_VAL
    export EDGEX_STARTUP_INTERVAL
fi

exec "$@"
