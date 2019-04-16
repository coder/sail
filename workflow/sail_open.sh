#!/bin/bash
set -e

sail open $(sail ls | cut -f1 -d" " | tail -n +2 | fzf --height 5)