#!/bin/bash
cd "$(dirname "$0")"
go build -o arts .
sudo cp arts /usr/bin/arts
