#!/bin/bash

go build -buildmode=plugin -gcflags="all=-N -l" -o transform.so transform.go
