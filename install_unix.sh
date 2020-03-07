#!/bin/bash

GO_IP="127.0.103.111"

grep -q "$GO_IP" /etc/hosts || ( echo "Adding to /etc/hosts..." ; echo "$GO_IP go" | sudo tee /etc/hosts)
