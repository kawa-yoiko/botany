#!/bin/sh

echo "Compiling $1"
SID=$1
LANG=$2
CODE=submissions/$SID/code.$LANG
BIN=submissions/$SID/bin

if [[ "$LANG" == "gcc.c" ]]; then
    gcc $CODE -O2 -I/var/botany/lib /var/botany/lib/bot.c -o $BIN 2>&1
elif [[ "$LANG" == "gcc.cpp" ]]; then
    g++ $CODE -O2 -I/var/botany/lib -x c /var/botany/lib/bot.c -o $BIN 2>&1
fi
