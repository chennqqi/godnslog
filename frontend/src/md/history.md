# update history

## v0.5.0

New Features:
1. GODNSLOG Can be used as standard name server whitch support `A,TXT,CNAME,MX` four kinds resolve types. Pan-domain name resolution is not supported.
2. Add backend translation

## v0.4.0

New Features:

1. Support [xip](http://xip.io/)
    10.0.0.1.godnslog.com   resolves to   10.0.0.1
    www.10.0.0.1.godnslog.com   resolves to   10.0.0.1
    mysite.10.0.0.1.godnslog.com   resolves to   10.0.0.1
    foo.bar.10.0.0.1.godnslog.com   resolves to   10.0.0.1
2. Add subcommands to start application
3. Add `resetpwd` subcommand for reset user's password

## v0.3.0

Add client sdk
Autobuild by using dockerhub

## v0.2.0

fix bugs
add docker support

## v0.1.0

init project