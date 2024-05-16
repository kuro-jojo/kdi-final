#!/bin/bash

# Export the content as an environment variable
export KDI_WEBAPP_WEP_API_ENDPOINT=$(cat /tmp/kdi_web_api_endpoint.txt)
export KDI_WEBAPP_MSAL_CLIENT_ID=$(cat /tmp/kdi_web_msal_client_id.txt)
export KDI_WEBAPP_MSAL_AUTHORITY=$(cat /tmp/kdi_web_msal_authority.txt)
export KDI_WEBAPP_MSAL_REDIRECT_URI=$(cat /tmp/kdi_web_msal_redirect_uri.txt)
export KDI_WEBAPP_MSAL_SCOPE=$(cat /tmp/kdi_web_msal_scope.txt)

# Remove the file
rm /tmp/kdi_*.txt

# Run the main command with exec to pass control
exec "$@"