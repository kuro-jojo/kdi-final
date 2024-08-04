#!/bin/bash

encoded_string=$(echo -n "" | base64 -w 0)
echo "Encoded string: $encoded_string"
