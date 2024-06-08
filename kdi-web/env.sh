#!/bin/bash

encoded_string=$(echo -n "http://kdi-webapp:8080" | base64 -w 0)
echo "Encoded string: $encoded_string"