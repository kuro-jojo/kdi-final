#!/bin/bash

encoded_string=$(echo -n "mongodb+srv://kdi:rnFiJaGZTHwKw8k0@kdi-cluster.mnpis8w.mongodb.net/?retryWrites=true&w=majority&appName=kdi-cluster&tls=true" | base64 -w 0)
echo "Encoded string: $encoded_string"