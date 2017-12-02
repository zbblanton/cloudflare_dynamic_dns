# Cloudflare Dynamic DNS

## Description
Small program that will automatically update your Cloudflare DNS with your dynamic public IP.

## Getting Started
Will update with more detailed instructions soon.

```
Rename config.json.example -> config.json
```

Create an API user on Cloudflare and get your Zone ID from the overview page.
Use this information to fill out the config.json file's cloudflare_api section.

No need to modify the public_ip_urls section.

You can change the interval (in minutes). It's default will check every 1 minute.

## Example
```
./cloudflare_dynamic_dns
```
