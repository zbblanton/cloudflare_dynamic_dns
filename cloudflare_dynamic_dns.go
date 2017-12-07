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
  "net/smtp"
)

type Cloudflare_api struct {
	Auth_email string
  Api_base_url string
  Api_key string
  Zone_id string
  Dns_record_name string
}

type Smtp_config struct {
  Host string
  Port string
  User string
  Pass string
  To string
  From string
  Enable bool
}

type Config_file struct{
  Cloudflare_api Cloudflare_api
  Auth_email string
  Api_key string
  Zone_id string
  Dns_record_name string
  Public_ip_urls []string
  Interval int
  Smtp Smtp_config
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

type Cf_error struct {
  Code int
  Message string
}

type Cf_data struct {
  Success bool
  Result []Cf_result
  Result_info Cf_result_info
  Errors []Cf_error
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

func sendmail(c Smtp_config, s string, m string) error {
  //Only send if smtp is enabled.
  if(c.Enable){
  	auth := smtp.PlainAuth("", c.User, c.Pass, c.Host)
  	to := []string{c.To}
  	msg := []byte("To: " + c.To + "\r\n" + "Subject: " + s + "\r\n" + "\r\n" + m + "\r\n")
    h := c.Host + ":" + c.Port
  	err := smtp.SendMail(h, auth, c.From, to, msg)
  	if err != nil {
  		return(err)
  	}
  }

	return nil
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

  //Check if success, print errors from api if not.
  if !r.Success {
    for _, e := range r.Errors {
      fmt.Printf("Error code %s: %s", e.Code, e.Message)
    }
    log.Fatal("Api call failed.")
  }

  if r.Result_info.Total_count < 1 {
    return Cf_result{}, fmt.Errorf("Cloudflare found no results for %s", dns_record_name)
  }
	return r.Result[0], nil
}

func (c Cloudflare_api) dns_update(id string, name string, ip string) error {
  api_url := c.Api_base_url + c.Zone_id + "/dns_records/" + id

  json_data := `
    {
      "type": "A",
      "name": "` + name + `",
      "content": "` + ip + `",
      "ttl": 1,
      "proxied": false
    }
  `

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

  r := Cf_data{}
  json.NewDecoder(resp.Body).Decode(&r)

  //Check if success, print errors from api if not.
  if !r.Success {
    for _, e := range r.Errors {
      fmt.Printf("Error code %d: %s\n", e.Code, e.Message)
    }
    return fmt.Errorf("Api call failed")
  }

  /*
  Marked to be deleted after testing.
  resp_data, err := ioutil.ReadAll(resp.Body)
  if err != nil {
      log.Fatal(err)
  }
  */

	return nil
}

func check_cf(cf_api Cloudflare_api, smtp Smtp_config, curr_ip string) error {
  fmt.Printf("Getting info for %s DNS record on Cloudflare.\n", cf_api.Dns_record_name)
  cf_info, err := cf_api.dns_record_info(cf_api.Dns_record_name)
  if(err != nil){
    fmt.Println("Not found: DNS Record will be added.")
    return fmt.Errorf("Adding a new record is not supported yet.")
  }

  cf_ip := cf_info.Content

  fmt.Printf("Current public IP is: %s\n", curr_ip)
  if(curr_ip != cf_ip){
    fmt.Println("Public IP has changed. Updating Cloudflare.")
    err := cf_api.dns_update(cf_info.Id, cf_api.Dns_record_name, curr_ip)
    if(err != nil){
      return err
    }
    fmt.Println("IP Updated.")

    //Send email
    s := "Public IP Changed to: " + curr_ip
    m := "Your Dynamic IP has changed to " + curr_ip + ". Cloudflare DNS has been updated."
    err = sendmail(smtp, s, m)
    if(err != nil){
      //TODO: Add code to print actual error.
      fmt.Println("Email could not be sent.")
    }
  } else {
    fmt.Println("No IP change.")
  }

  return nil
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

  for {
    curr_ip := get_public_ip(config_data.Public_ip_urls[0])
    check_cf(cf_api, config_data.Smtp, curr_ip)
    time.Sleep(time.Duration(config_data.Interval * 60) * time.Second)
  }
}
