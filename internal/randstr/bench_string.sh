#!/usr/bin/env bash

go test -bench=String -run=^$ -cpuprofile=/tmp/randstr.cpuprof -o /tmp/randstr.test
