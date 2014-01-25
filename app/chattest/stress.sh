#!/bin/bash

for i in {1..$1}
do
    sleep 1s
    ./chattest chat.winterwest.net:8900 $i 1000 > info$i.log 2> err$i.log &
done

