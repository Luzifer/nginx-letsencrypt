# Luzifer / nginx-letsencrypt

`nginx-letsencrypt` is a wrapper tool around nginx which takes care of automated configuration reloads and - as the name states - also ensures all nginx `server`s do have valid certificates.

## Features

- Automatically read nginx configuration and determine FQDNs to request certificates for
- Bundle certificates under each TLD to minimize certificate requests
- Trigger reload of nginx when configuration or certificates have been changed

## Docker container

The container is intended to be used for example with [`docker-compose`](https://docs.docker.com/compose/). You can start your containers with your compose file and add this container to them, putting the nginx configuration next to the container.

For an example see the `example` directory.

----

![project status](https://d2o84fseuhwkxk.cloudfront.net/nginx-letsencrypt.svg)
