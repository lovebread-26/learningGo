#!/bin/bash

usage() {
  echo "./request [reload|print|clear|custid]"
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
	curl "http://127.0.0.1:8080/blacklist?custid=$2&token=395e1cfd662b49ff92bc37b6c15cab62"
  ;;
  *)
  usage
  ;;
esac
