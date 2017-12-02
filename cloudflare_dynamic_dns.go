package main

import(
  "fmt"
  "os"
  "log"
  "net/http"
  //"io"
  "encoding/json"
  "strings"
  "io/ioutil"
  "time"
)

type Cloudflare_api struct {
	Auth_email string
  Api_base_url string
  Api_key string
  Zone_id string
  Dns_record_name string
}

type Config_file struct{
  Cloudflare_api Cloudflare_api
  Auth_email string
  Api_key string
  Zone_id string
  Dns_record_name string
  Public_ip_urls []string
  Interval int
}

type Cf_result struct {
  Id string
  Type string
  Name string
  Content string
}

type Cf_result_info struct {
  Total_count int
}

type Cf_data struct {
  Success bool
  Result []Cf_result
  Result_info Cf_result_info
}

func get_public_ip(url string) string {
  client := &http.Client{}
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    log.Fatal(err)
  }
  resp, err := client.Do(req)
  if err != nil {
    log.Fatal(err)
  }
  defer resp.Body.Close() //Close the resp body when finished

  //Use ioutil to read the byte slice from resp body, use this to return as a string, then trim the newline
  r, err := ioutil.ReadAll(resp.Body)
  if err != nil {
      log.Fatal(err)
  }
  s := string(r)
  s = strings.TrimSuffix(s, "\n")

  return s
}

func (c Cloudflare_api) dns_record_info(dns_record_name string) (Cf_result, error) {
  api_url := c.Api_base_url + c.Zone_id + "/dns_records?type=A&name=" + dns_record_name

  client := &http.Client{}
  req, err := http.NewRequest("GET", api_url, nil)
  if err != nil {
    log.Fatal(err)
  }
  req.Header.Add("X-Auth-Email", c.Auth_email)
  req.Header.Add("X-Auth-Key", c.Api_key)
  req.Header.Add("Content-type", "application/json")
  resp, err := client.Do(req)
  if err != nil {
    log.Fatal(err)
  }
  defer resp.Body.Close() //Close the resp body when finished

  r := Cf_data{}
  json.NewDecoder(resp.Body).Decode(&r)

  if r.Result_info.Total_count < 1 {
    return Cf_result{}, fmt.Errorf("Cloudflare found no results for %s", dns_record_name)
  }
	return r.Result[0], nil
}

func (c Cloudflare_api) dns_update(id string, name string, ip string) string {
  api_url := c.Api_base_url + c.Zone_id + "/dns_records/" + id
  fmt.Println(ip)
  json_data := `
    {
      "type": "A",
      "name": "` + name + `",
      "content": "` + ip + `",
      "ttl": 1,
      "proxied": false
    }
  `
  fmt.Println(json_data)

  client := &http.Client{}
  req, err := http.NewRequest("PUT", api_url, strings.NewReader(json_data))
  if err != nil {
    log.Fatal(err)
  }
  req.Header.Add("X-Auth-Email", c.Auth_email)
  req.Header.Add("X-Auth-Key", c.Api_key)
  req.Header.Add("Content-type", "application/json")
  resp, err := client.Do(req)
  if err != nil {
    log.Fatal(err)
  }
  defer resp.Body.Close() //Close the resp body when finished


  resp_data, err := ioutil.ReadAll(resp.Body)
  if err != nil {
      log.Fatal(err)
  }
  fmt.Printf(string(resp_data))



	return "resp.Body"
}

func main() {
  file, err := os.Open("config.json")
  if err != nil {
    fmt.Println("Did you rename config.json.example to config.json? :) ")
  	log.Fatal(err)
  }
  config_data := Config_file{}
  json.NewDecoder(file).Decode(&config_data)
  file.Close()

  cf_api := Cloudflare_api{
  	config_data.Cloudflare_api.Auth_email,
    "https://api.cloudflare.com/client/v4/zones/",
    config_data.Cloudflare_api.Api_key,
    config_data.Cloudflare_api.Zone_id,
    config_data.Cloudflare_api.Dns_record_name}

  fmt.Printf("Searching for %s DNS record on Cloudflare\n", cf_api.Dns_record_name)
  cf_info, err := cf_api.dns_record_info(cf_api.Dns_record_name)
  if(err != nil){
    fmt.Printf("Not found: Record will be added.\n")
    log.Fatal("Adding a record is not supported yet.")
  } else {
    fmt.Printf("Found DNS record.\n")
  }
  id := cf_info.Id
  cf_ip := cf_info.Content
  for range time.NewTicker(time.Duration(config_data.Interval * 60) * time.Second).C {
    curr_ip := get_public_ip(config_data.Public_ip_urls[0])
    fmt.Printf("Current public IP is: %s\n", curr_ip)
    if(curr_ip != cf_ip){
      fmt.Printf("Public IP has changed. Updating Cloudflare\n")
      cf_api.dns_update(id, cf_api.Dns_record_name, curr_ip)
      cf_ip = curr_ip
    }
  }
}
