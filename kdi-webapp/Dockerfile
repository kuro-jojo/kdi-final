# Use official node image as the base image
FROM node:latest as build

# Set the working directory
WORKDIR /usr/local/app

# Copy the package.json to the working directory
COPY package*.json ./

# Run npm clean install
RUN npm ci

# Add the source code to app
COPY ./ /usr/local/app/

# Load secrets

RUN --mount=type=secret,id=kdi_web_api_endpoint cat /run/secrets/kdi_web_api_endpoint  > /tmp/kdi_web_api_endpoint.txt
RUN --mount=type=secret,id=kdi_web_msal_client_id cat /run/secrets/kdi_web_msal_client_id  > /tmp/kdi_web_msal_client_id.txt
RUN --mount=type=secret,id=kdi_web_msal_authority cat /run/secrets/kdi_web_msal_authority  > /tmp/kdi_web_msal_authority.txt
RUN --mount=type=secret,id=kdi_web_msal_redirect_uri cat /run/secrets/kdi_web_msal_redirect_uri  > /tmp/kdi_web_msal_redirect_uri.txt
RUN --mount=type=secret,id=kdi_web_msal_scope cat /run/secrets/kdi_web_msal_scope  > /tmp/kdi_web_msal_scope.txt 

# Load secrets
# Copy the script into the image
COPY set-env.sh /set-env.sh

# Make the script executable
RUN chmod +x /set-env.sh

# Generate the build of the application
RUN  /bin/bash -c "source /set-env.sh && npm run build --configuration=production"


# Stage 2: Serve app with nginx server

# Use official nginx image as the base image
FROM nginxinc/nginx-unprivileged

COPY ./nginx.conf /etc/nginx/conf.d/default.conf

# Copy the build output to replace the default nginx contents.
COPY --from=build /usr/local/app/dist/kdi-webapp /usr/share/nginx/html
# RUN usermod -u 1001 nginx && chown -R 1001:0 /var/cache/nginx 
    # Expose port 80
EXPOSE 8080