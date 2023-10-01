#!/bin/bash

sudo iptables -i lo -p icmp -A INPUT -j DROP
