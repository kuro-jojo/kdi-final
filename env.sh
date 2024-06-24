#!/bin/bash

encoded_string=$(echo -n "87zrgrgzz6c7ae1fad8f1808debdee9816a4fd4b3f420e50b2dd5ecef19d25e26f2e9a2727920cb5ed0d1831e4a8fce96707bfa60bf7c51649175237cgrzgregza09766e6dab379" | base64 -w 0)
echo "Encoded string: $encoded_string"
