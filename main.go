package main

import(
	"fmt"
	"strconv"
	"errors"
	"strings"
	"io"
	"net"
	"srun_smbu/hash"
	"srun_smbu/model"
	"net/url"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
	"flag"
)
var username string	//账号
var baseAddr string //登录地址
var password string //密码
var nwip	string 	//网卡ip

var err error
var client *http.Client
const (
	challengeUrl = "/cgi-bin/get_challenge"
	portalUrl    = "/cgi-bin/srun_portal"
)

func genCallback() string {
	return fmt.Sprintf("jsonp%d", int(time.Now().Unix()))
}
func DoRequest(url string, params url.Values) (*http.Response, error) {
	// add callback
	params.Add("callback", genCallback())
	params.Add("_", fmt.Sprint(time.Now().UnixNano()))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func GetJson(url string, data url.Values, res interface{}) (err error) {
	resp, err := DoRequest(url, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	rawStr := string(raw)
	// cut jsonp
	start := strings.Index(rawStr, "(")
	end := strings.LastIndex(rawStr, ")")
	if start == -1 && end == -1 {
		log.Debug("raw response:", rawStr)
		return errors.New("error-parse")
	}
	dt := string(raw)[start+1 : end]

	return json.Unmarshal([]byte(dt), &res)
}

func get(addr string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodGet, addr, nil)
	return client.Do(req)
}
//get acid
func Prepare() (int, error) {
	first, err := get(baseAddr)
	if err != nil {
		return 1, err
	}
	second, err := get(first.Header.Get("Location"))
	if err != nil {
		return 1, err
	}
	target := second.Header.Get("location")
	query, _ := url.Parse(baseAddr + target)
	return strconv.Atoi(query.Query().Get("ac_id"))
}
func getChallenge() (res model.ChallengeResp, err error) {
	qc := model.ChallengeVal(username)
	err = GetJson(baseAddr+challengeUrl, qc, &res)
	return
}
func initHttpClient(localIP string) (*http.Client, error) {
	parsedIP := net.ParseIP(localIP)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", localIP)
	}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				LocalAddr: &net.TCPAddr{IP: parsedIP},
			}).DialContext,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // disable auto redirect
		},
	}
	return client, nil
}

func main(){
	flag.StringVar(&username,"user","","username")
	flag.StringVar(&password,"pass","","password")
	flag.StringVar(&baseAddr,"addr","http://172.20.5.18","base address")
	flag.StringVar(&nwip,"nwip","","network ip")
	flag.Parse()
	log.SetLevel(log.DebugLevel)
	if(nwip==""){
		log.Fatalln("No network card ip is specified.")
	}
	client,err=initHttpClient(nwip)
	if err!=nil{
		log.Fatalln("prepare failed while initHttpClient "+err.Error())
	}

	acid, err := Prepare() //get acid
	if err != nil {
		log.Fatalln("prepare error:", err)
	}
	log.Println("getacid",acid)
	formLogin := model.LoginVal(username,password,acid)
	rc, err := getChallenge()
	if err != nil {
		log.Debug("get challenge error:", err)
		return
	}
	log.Println(rc.Challenge)
	token := rc.Challenge
	ip := rc.ClientIp
	formLogin.Set("ip", ip)
	formLogin.Set("info", hash.GenInfo(formLogin, token))
	formLogin.Set("password", hash.PwdHmd5("", token))
	formLogin.Set("chksum", hash.Checksum(formLogin, token))
	ra := model.ActionResp{}
	if err = GetJson(baseAddr+portalUrl, formLogin, &ra); err != nil {
		log.Println("request error", err)
		return
	}
	log.Println(ra.Res)
	if ra.Res != "ok" {
		log.Println("response msg is not 'ok'")
		if strings.Contains(ra.ErrorMsg, "Arrearage users") {
			err = errors.New("已欠费")
		} else {
			fmt.Println(ra)
		}
		return
	}

}