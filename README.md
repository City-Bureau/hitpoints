# Hitpoints

![Build status](https://github.com/City-Bureau/hitpoints/workflows/CI/badge.svg)

Minimal tool for counting page hits on embedded content. Inspired by [`pixel-ping`](https://github.com/documentcloud/pixel-ping).

## Priorities

* Simple to setup on a single server
* Speed, not slowing anything down
* Writes everything to static file storage for easier archiving
* Prioritizes simplicity over absolute accuracy

## Deployment

Deployment setups for AWS and Azure using [Terraform](https://www.terraform.io/) are available in the [`deploy/`](./deploy) directory. In general, they run the following steps:

* Setup a storage container/bucket for the outputs
* Provision a small (500MB RAM) server
* Setup a swapfile
* Configure security rules for inbound traffic to only allow inbound connections for HTTP, HTTPS, and SSH (with provided key)
* Copy a release binary
* Setup `systemd` service for running the service

You'll need to point the domain you want to use to the IP that comes out of outputs before the SSL certificate can be created.

## Using

Include snippet on target page with Javascript

```html
<script type="text/javascript" src="https://{DOMAIN}/hitpoints.js" async="true"></script>
```

or with an image tag

```html
<img src="https://{DOMAIN}" width="1" height="1" alt="" />
```
