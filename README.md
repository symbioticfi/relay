# Offchain Middleware

## Overview

Offchain Middleware is a peer-to-peer service designed to collect and aggregate BLS signatures from validators, form validator sets (valsets), and post the aggregated signatures to on-chain middleware contracts. This service facilitates efficient signature collection and aggregation in a decentralized manner.

## Repo init

This repo uses git-lsf, so make sure to install it first:

```bash
brew install git-lfs
git lfs install
git lfs pull
```

Then check that file content are downloaded

```bash
cat circuit/circuit_10.r1cs
```

## Commands

The application supports two commands:
Signer1 + Aggregator + Commitor:

```bash
middleware_offchain --master-address 0xF91E4B4166AD3eafDE95FeB6402560FCAb881690 --rpc-url http://127.0.0.1:8545 --log-level debug --secret-key 87191036493798670866484781455694320176667203290824056510541300741498740913410 --signer true --aggregator true --committer true --http-listen :8081
```

Signer2

```bash
middleware_offchain --master-address 0xF91E4B4166AD3eafDE95FeB6402560FCAb881690 --rpc-url http://127.0.0.1:8545 --log-level debug --secret-key 11008377096554045051122023680185802911050337017631086444859313200352654461863 --signer true --http-listen :8082
```

Signer3

```bash
middleware_offchain --master-address 0xF91E4B4166AD3eafDE95FeB6402560FCAb881690 --rpc-url http://127.0.0.1:8545 --log-level debug --secret-key 26972876870930381973856869753776124637336739336929668162870464864826929175089 --signer true --http-listen :8083
```