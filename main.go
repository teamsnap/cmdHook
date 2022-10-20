package main


import (
    "net/http"
    "os/exec"
    "github.com/gin-gonic/gin"
    "os"
	"net/url"
    "encoding/json"
	"bytes"
	"log"
	"fmt"
	"io/ioutil"
)

var credData = url.Values{
	"audience":       {"https://api.cloud.armory.io"},
	"grant_type": {"client_credentials"},
	"client_secret": {os.Getenv("client-secret")},
	"client_id": {os.Getenv("client-id")},
}

type callback_data struct {
    success    bool  `json:"success"`
    mdMessage  string  `json:"mdMessage"`
}

type auth_data struct {
    access_token  string
	scope	string
	expires_in int
	token_type	string
}

// getAlbums responds with the list of all albums as JSON.
func runCmd(c *gin.Context) {
    fmt.Println("request recieved")
	cmd:=c.Query("cmd")
	arg:=c.Query("arg")
	callbackURL:=c.Query("callbackURL")
	out, err := exec.Command(cmd, arg).Output()
	message:=""
	success:=true
    fmt.Println(out)
	if err!=nil {
		message=err.Error()
		c.IndentedJSON(http.StatusInternalServerError,err.Error())
		success=false
	} else {
		c.IndentedJSON(http.StatusOK,string(out[:]))
		message=string(out[:])
	}
	
	token:=auth()
    fmt.Println("Authorized")
	callback(token, callbackURL,success,message)
}

func callback(token string,callbackURL string, success bool, messae string){
	data := &callback_data{true, "message"}
	serialized, err :=json.Marshal(data)

    var bearer = "Bearer " + token
	client := &http.Client{}
	
    fmt.Println("posting: "+bearer)
    fmt.Println(callbackURL)
	req,err := http.NewRequest("POST",callbackURL,bytes.NewBuffer(serialized))
    req.Header.Add("Authorization", bearer)
	resp, err := client.Do(req)

    if err != nil {
        log.Fatal(err)
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(body))
}

func auth() string{
    resp, err := http.PostForm("https://auth.cloud.armory.io/oauth/token",credData)

    if err != nil {
        log.Fatal(err)
    }

    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(body))
	var access_token map[string]interface{}
	err = json.Unmarshal([]byte(body),&access_token)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("unmarshalled:")
    fmt.Println(access_token["access_token"].(string))
	return access_token["access_token"].(string)
}

func main() {
    router := gin.Default()
    router.GET("/cmd", runCmd)

    fmt.Println("starting")
    //router.Run("localhost:8080")
    router.Run()
}