#!/usr/bin/kumachan

module Hello;

do Stdout
   | println 'Hello World'
   | then exit 0
   | catch exit 1;

