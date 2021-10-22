#!/usr/bin/env sh

max_retry=30

for var in "$@"
do
    host=${var%:*}
    port=${var##*:}

    n=0

    until [ $n -ge $max_retry ]
    do
      nc -z $host $port && break
      sleep 1
      echo "[$n] waiting for $host:$port"
      n=$(($n + 1))
    done
    if [ "$n" -eq "$max_retry" ]; then
        echo ">> $host:$port is still not ready 😔"
        exit 1
    fi
done