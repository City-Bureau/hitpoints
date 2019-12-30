# Hitpoints

![Build status](https://github.com/City-Bureau/hitpoints/workflows/CI/badge.svg)

Simple setup for tracking page hits across multiple domains through an embedded 1x1 pixel GIF.

## Priorities

* Simple to setup on a single server
* Speed, not slowing anything down
* Writes everything to static file storage for easier archiving
* Prioritizes simplicity over absolute accuracy

## Deployment

Include snippet on target page with Javascript

```html
<script type="text/javascript" src="https://{DOMAIN}/hitpoints.js" async="true"></script>
```

or with an image tag

```html
<img src="https://{DOMAIN}" width="1" height="1" alt="" />
```

## To-Do

- [ ] Terraform setup for deploying to AWS, Azure
- [ ] Basic admin dashboard (authentication or setup for HTTP basic)
