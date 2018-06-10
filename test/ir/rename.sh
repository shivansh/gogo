#!/bin/bash

for f in *.s; do
    file=${f%.s}
    s=${file##*/}
    mv $f "$s.asm"
done
