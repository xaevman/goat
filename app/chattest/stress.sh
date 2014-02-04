#!/bin/bash

for (( i = 0; i < $1; i++ ))
do
    sleep 1s
    chattest chat.winterwest.net:8900 $i $2 > info$i.log 2> err$i.log &
done
