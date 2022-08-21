# go-forward-ssdp

SSDP packet forwarder

## Introduction

Like many people, I segment my home network such that IoT devices are on a different VLAN than
my phone and laptop. Unfortunately, this breaks things like the Roku app because SSDP isn't meant
to cross network boundaries.

I run this code on my OPNsense router to restore sanity to my home network. It works well with
IPv4 traffic, but it has not been fully tested with IPv6. (Don't get me started on the irony of
most IoT devices lacking support for IPv6.)

I'm not a network or protocol engineer, so don't assume this code is correct or bug-free.
