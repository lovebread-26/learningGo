#!/bin/bash

usage() {
  echo "./request [reload|print|clear|custid|bind]"
}

case "$1" in
  "print")
	curl "http://127.0.0.1:8080/blacklist?print=1"
  ;;
  "reload")
	curl "http://127.0.0.1:8080/blacklist?reload=1"
  ;;
  "clear")
    curl "http://127.0.0.1:8080/blacklist?clear=1"
  ;;
  "custid")
	curl  "http://127.0.0.1:8080/blacklist?custid=$2&token=395e1cfd662b49ff92bc37b6c15cab6$3" -T requestbody.json
  ;;
  "token")
	curl "http://127.0.0.1:8080/blacklist?token=$2"
  ;;
  "bind")
	curl "http://127.0.0.1:3190/gw/bind/insert/binding/insertWorkAXB?access_token=395e1cfd662b49ff92bc37b6c15cab6$2" -T requestbody.json
  ;;
  "burst")
	< burst.txt xargs -r -L 1 -P 10 curl
  ;;
  "bprint")
	curl "http://127.0.0.1:3190/proxy/gw/bind/insert/binding/insertWorkAXB?access_token=395e1cfd662b49ff92bc37b6c15cab62&blacklistprint=1"
  ;;
  "breload")
	curl "http://127.0.0.1:3190/proxy/gw/bind/insert/binding/insertWorkAXB?access_token=395e1cfd662b49ff92bc37b6c15cab62&blacklistreload=1"
  ;;
  "bclear")
	curl "http://127.0.0.1:3190/proxy/gw/bind/insert/binding/insertWorkAXB?access_token=395e1cfd662b49ff92bc37b6c15cab62&blacklistclear=1"
  ;;
  *)
  usage
  ;;
esac
