/**
    @Author : Loneyers
    @Date : 2020/8/24
    @FileName : main
**/

package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	engine	string
	query string
	result	[]string
	access_token string
	zoomeye_login_api = "https://api.zoomeye.org/user/login"
)

type FofaConfig struct {
	Fofa struct {
		Email        string   `yaml:"email"`
		Key      string   `yaml:"key"`
	}
}
type ZoomeyeConfig struct {
	Zoomeye struct {
		Email        string   `yaml:"email"`
		Password      string   `yaml:"password"`
	}
}
func init(){
	flag.StringVar(&engine,"engine","","zoomeye")
	flag.StringVar(&query,"query","","query")
}
func getToken() string{
	config := new(ZoomeyeConfig)
	file,err:=ioutil.ReadFile("config.yaml")
	if err!=nil{
		log.Println(err)
	}
	err =yaml.Unmarshal(file,&config)
	if err!=nil{
		log.Println(err)
	}
	post := fmt.Sprintf(`{"username":"%s","password":"%s"}`,config.Zoomeye.Email,config.Zoomeye.Password)
	req,err:=http.NewRequest("POST",zoomeye_login_api,bytes.NewBuffer([]byte(post)))
	req.Header.Add("Content-Type","application/x-www-form-urlencoded")
	client:=&http.Client{}
	resp,err:=client.Do(req)
	if err!=nil{
		panic(err)
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err!=nil{
		log.Fatal(err)
	}
	access_token := gjson.Get(string(result),"access_token")
	return access_token.String()
}

func fofa(q string){
	config := new(FofaConfig)
	file,err:=ioutil.ReadFile("config.yaml")
	if err!=nil{
		log.Println(err)
	}
	err =yaml.Unmarshal(file,&config)
	if err!=nil{
		log.Println(err)
	}
	if config.Fofa.Email == "" ||config.Fofa.Key == ""{
		fmt.Println("email or key is empty.")
		os.Exit(0)
	}
	base64q := base64.StdEncoding.EncodeToString([]byte(q))
	url := fmt.Sprintf("https://fofa.so/api/v1/search/all?email=%s&key=%s&qbase64=%s&size=100&page=1&full=true",config.Fofa.Email,config.Fofa.Key,base64q)
	resp,err := http.Get(url)
	if err!=nil{
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!=nil{
		log.Fatal(err)
	}
	for _, v := range gjson.Get(string(body), `results.#.0`).Array() {
		result = append(result,v.String())
		fmt.Println(v.String())
	}

}

func zoomeye(query string){
	url := fmt.Sprintf("https://api.zoomeye.org/web/search?query=%s&page=1&facets=app,os",query)
	fmt.Println(url)
	req, err := http.NewRequest("GET", url,nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "JWT "+getToken())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body,_ := ioutil.ReadAll(resp.Body)
	for _, v := range gjson.Get(string(body), `matches.#.webapp.0.url`).Array() {
		fmt.Println(v.String())
		result = append(result,v.String())
	}
}
func main(){
	flag.Parse()
	switch engine {
	case "fofa":
		fofa(query)
	case "zoomeye":
		zoomeye(query)
	default:
		fmt.Println("./search -engine fofa -query app=\"IIS\"\n./search -engine zoomeye -query app:Solr")
	}
}
