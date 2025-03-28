#!/bin/bash

# wait for minio server
sleep 5

mc alias set local http://localhost:9000 minioadmin minioadmin

buckets="vultisig-verifier vultisig-plugin"

for bucket in $buckets; do
    mc mb local/$bucket
done
