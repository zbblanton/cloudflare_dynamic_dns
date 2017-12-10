# Cloudflare Dynamic DNS

## Description
Small program that will automatically update your Cloudflare DNS with your dynamic public IP.

## Getting Started
Compiling from source:
```
go build
```
Or download the latest binary from the [Releases](https://github.com/zbblanton/cloudflare_dynamic_dns/releases). 

```
Rename config.json.example -> config.json
```

Create an API user on Cloudflare and get your Zone ID from the overview page.
Use this information to fill out the config.json file's cloudflare_api section.

Optional: Configure the SMTP settings and set "enable" to "true".

No need to modify the public_ip_urls section.

You can change the interval (in minutes). It's default will check every 1 minute.

## Examples
Run with default everything:
```
./cloudflare_dynamic_dns
```

Run as a cronjob:
```
*/1 * * * * ./path/to/cloudflare_dynamic_dns --cron=true --config=/path/to/config.json --log=true --log_path=/path/to/cloudflare_dynamic_dns.log
```

Run as background task:
```
nohup ./path/to/cloudflare_dynamic_dns --config=/path/to/config.json --log=true --log_path=/path/to/cloudflare_dynamic_dns.log &
```
