# knotenalarm

Dies ist der Freifunk Hamburg [knotenalarm](https://twitter.com/knotenalarm).

## Dependencies

* https://github.com/ChimeraCoder/anaconda
* https://github.com/scalingdata/gcfg

## Installation & Usage

1. create an application on twitter https://apps.twitter.com/app/new
2. edit configfile and `mv myconfig.gcfg.sample myconfig.gcfg`
3. set up a cronjob `37 * * * * /usr/bin/go run /foo/knotenalarm.go`, don't forget the GOPATH
