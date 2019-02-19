#!/bin/bash


repeat() { printf "$1"'%.s' $(eval "echo {1.."$(($2))"}"); }


while read line; do
    t=$(echo "$line" | sed -r 's/[^:]*:([0-9]+):([0-9]+): (.*)/\1::\2::\3/')
    n=$(echo "$t" | sed 's/::.*//')
    x=$(echo "$t" | sed -r 's/[^:]*:://')
    c=$(echo "$x" | sed 's/::.*//')
    err=$(echo "$x" | sed 's/[^:]*:://')
    if [[ "$n" =~ ^[0-9]+$ ]]; then
        echo '--------------------------'
        spaces=$(repeat '\ ' $(expr "$c" - 1))
        bar="${spaces}$(echo -e '\e[1m\e[31m^\e[0m')"
        highlight=$(echo -e '\e[31m')
        end=$(echo -e '\e[0m')
        echo "line ${n}, column ${c}: ${err}"
        cat build/kumago.go | sed -n "$(expr ${n} - 1),$(expr ${n} + 1)p;" | sed "2s/.*/${highlight}&${end}/;3i${bar}"
        echo '--------------------------'
        echo ''
    else
        echo "$line"
    fi
done
