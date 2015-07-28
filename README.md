# knotenalarm

Dies ist der Freifunk Hamburg [knotenalarm](https://twitter.com/knotenalarm).

## Dependencies

* https://github.com/ChimeraCoder/anaconda
* https://github.com/scalingdata/gcfg

## Installation & Usage

1. create an application on twitter https://apps.twitter.com/app/new
2. edit configfile and `mv myconfig.gcfg.sample myconfig.gcfg`
3. Use `go build knotenalarm.go` and put the resulting binary in your `$PATH`.
4. set up a cronjob
