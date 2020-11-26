# kthxbye

kthxbye is a simple daemon that works as a sidecar for
[Prometheus Alertmanager](https://github.com/prometheus/alertmanager) and will automatically extend expiring silences.
The goal of this project is to provide a simple way of acknowledging alerts,
which is currently not possible with Alertmanager itself
[see this issue](https://github.com/prometheus/alertmanager/issues/1860).

## Current acknowledgment workflow with Alertmanager

Currently when a new alert fires in Alertmanager there are 2 options:

- Leave the alert in active state while you work on resolving it
- Silence this alert for some duration

This works well in small environments but can cause problems with big teams:

- If you leave alert in active state you need to communicate somehow that
  you are working on it, otherwise someone else on the team might start
  working on it too, or (worse) nobody will work on it assuming that someone
  else already does.
- If you silence this alert you need to pick the correct duration.
  If the duration is too short you might need to re-silence it again as it expires, which can be noisy with a lot of Alertmanager users.
  If the duration is too long you need to remember to expire it after the issue
  was resolved, otherwise it might mask other problems or issue reoccurring
  soon.

There are tools to better manage alert ownership like PagerDuty or Opsgenie,
which can help to avoid this problem, but they require sending all alerts
to external escalation system.

## How it works

kthxbye will continuously extend silences that are about to expire but are
matching firing alerts. Silences will be allowed to expire only if they don't
match any alert.

- A new alert starts to fire in Alertmanager
- User creates a silence for it with a comment that beings with predefined
  prefix and short duration
- kthxbye will continuously poll alerts and silences from Alertmanager:
  - Get the list of all silences
  - Get the list of all silenced alerts
  - Find all silences where comments starts with predefined prefix and are
    expiring soon
  - For every such silence:
    - Check if the silence matches any currently firing alerts
      - If yes then silence duration will be extended
      - If no then silence will be allowed to expire

This allows to silence an alert without worrying about picking correct duration
for the silence, so you effectively silence a specific indecent rather than
the alert.

## Building binaries

Have the most recent Go version and compile it using the usual `go build`
command:

```shell
go build ./...
```

## Docker images

Images are built automatically for:

- release tags in git - ghcr.io/prymitive/kthxbye:vX.Y.Z
- master branch commits - ghcr.io/prymitive/kthxbye:latest

## Usage

Start kthxbye and pass the address of Alertmanager you want it to manage.

```shell
kthxbye -alertmanager.uri http://alertmanager.example.com:9093
```

By default kthxbye will only extend silences with comment starting with `ACK!`,
you can set a custom prefix with:

```shell
kthxbye -extend-with-prefix "MY CUSTOM PREFIX"
```

Be sure to set `-extend-if-expiring-in` and `-extend-by` flags that match your
environment.
`-extend-if-expiring-in` tells kthxbye when to extend a silence, if you set it
to `6m` then it will extend all silences if they would expire in the next
6 minutes. Setting it to `90s` would tell kthxbye to extend silences expiring
in the next 90 seconds.
`-extend-by` tells kthxbye how much duration should be added to a silence when
it's extended. Setting it to `10m` would tell kthxbye that exteding a silence
would update it to expire 10 minutes from now.
`-max-duration` can be used to limit maximum duration a silence can reach, if
set to a non-zero value it will let kthxbye to extend silence duration only if
total duration (from the initial start time to the current expiry time) is less
than `-max-duration` value.

Note: duration values must be compatible with Go
[ParseDuration](https://golang.org/pkg/time/#ParseDuration) and so currently
valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

By default kthxbye will wake up and inspect silences every minute, you can
customize it by passing `-interval` flag. Setting it to `30s` would tell kthxbye
to wake up every 30 seconds.

It's recommended to run a single kthxbye instance per Alertmanager cluster.
If there are multiple instances running there will be a risk of races when
updating silences - two (or more) kthxbye instances might try to update the
same silence and, because Alertmanager expires the silence that's being updated
and creates a new silence with new values, it can cause updated silence to be
forked into two new silences.
kthxbye can tolerate some downtime, with default settings it only considers
silences with less than 5 minutes left, which means that it only needs to
run once during those 5 minutes to update such silence. This applies to each
silence individually, and they are unlikely to expire in the same time, so in
reality more than one execution per 5 minutes is required, but it means that
small amount of downtime that might be needed to restart kthxbye on another
k8s/mesos node (for cloud deployments) if the primary crashes is likely to be
invisible to the user.
