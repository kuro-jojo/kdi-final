#!/bin/bash

encoded_string=$(echo -n "https://kdi-webapp-kuro08-dev.apps.sandbox-m3.1530.p1.openshiftapps.com" | base64 -w 0)
echo "Encoded string: $encoded_string"