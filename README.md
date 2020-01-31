# Hitpoints

![Build status](https://github.com/City-Bureau/hitpoints/workflows/CI/badge.svg)

Minimal tool for counting page hits on embedded content. Inspired by [`pixel-ping`](https://github.com/documentcloud/pixel-ping). It's designed to be fast and simple by running a Go app on a single 500MB RAM server and writing all output to static file storage.

## Setup

Install dependencies and run a development server at `http://localhost:8080` with:

```bash
make install
make start
```

## Deployment

Deployment setups for AWS and Azure using [Terraform](https://www.terraform.io/) are available in the [`deploy/`](./deploy) directory. You'll need to create a release with `make release` before deploying.

Terraform will create static file storage, create and provision a server and configure network security rules. It will output the public IP address of the server when it's finished, and you'll need to update your DNS with an A record pointing to that IP.

## Usage

Include snippet on target page with JavaScript:

```html
<script type="text/javascript" src="https://{DOMAIN}/hitpoints.js" async="true"></script>
```

or with an image tag (`hitpoints.gif` isn't required, all endpoints other than `hitpoints.js` will return the pixel GIF):

```html
<img src="https://{DOMAIN}/hitpoints.gif" width="1" height="1" alt="" />
```
