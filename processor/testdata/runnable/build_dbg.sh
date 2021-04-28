#!/bin/bash

go build -buildmode=plugin -gcflags="all=-N -l" -o runnable.so runnable.go
