#!/bin/bash
rm blockchain
rm *.db

go build -o blockchain.exe *.go
./blockchain
