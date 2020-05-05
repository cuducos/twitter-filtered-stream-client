package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

const tokenUrl = "https://api.twitter.com/oauth2/token"
const rulesUrl = "https://api.twitter.com/labs/1/tweets/stream/filter/rules"
const streamUrl = "https://api.twitter.com/labs/1/tweets/stream/filter"

func crashIf(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

type Tweet struct {
	Id string
}

type TweetData struct {
	Data Tweet
}

func parseTweet(b []byte) (Tweet, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	var t TweetData
	for {
		if err := d.Decode(&t); err == io.EOF {
			break
		} else if err != nil {
			return Tweet{}, err
		}
	}
	return t.Data, nil
}

type Rule struct {
	Id    string
	Value string
}

type RuleData struct {
	Data []Rule
}

func parseRules(b []byte) []Rule {
	d := json.NewDecoder(bytes.NewReader(b))
	var r RuleData
	for {
		if err := d.Decode(&r); err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Error parsing this JSON:\n%s\n%s", string(b), err.Error())
		}
	}
	return r.Data
}

type Token struct {
	AccessToken string `json:"access_token"`
}

func parseToken(b []byte) Token {
	d := json.NewDecoder(bytes.NewReader(b))
	var t Token
	for {
		if err := d.Decode(&t); err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Error parsing this JSON:\n%b\n%s", b, err.Error())
		}
	}
	return t
}

func apiRequest(method string, url string, token string, body string) []byte {
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	crashIf(err)

	client := &http.Client{}
	req.Header.Set("Content-type", "application/json")
	if a := os.Getenv("TWITTER_APP_NAME"); a != "" {
		req.Header.Set("User-agent", a)
	}

	if token == "" {
		req.SetBasicAuth(os.Getenv("TWITTER_API_KEY"), os.Getenv("TWITTER_API_SECRET"))
		req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := client.Do(req)
	crashIf(err)
	defer res.Body.Close()

	c, err := ioutil.ReadAll(res.Body)
	crashIf(err)
	return c
}

func getToken() string {
	if t := os.Getenv("TWITTER_ACCESS_TOKEN"); t != "" {
		return t
	}

	b := apiRequest("POST", tokenUrl, "", "grant_type=client_credentials")
	return parseToken(b).AccessToken
}

func getRules() []byte {
	return apiRequest("GET", rulesUrl, getToken(), "")
}

func deleteRules() []byte {
	var i []string
	for _, r := range parseRules(getRules()) {
		i = append(i, fmt.Sprintf(`"%s"`, r.Id))
	}
	d := fmt.Sprintf(`{"delete": {"ids": [%s]}}`, strings.Join(i, ","))
	return apiRequest("POST", rulesUrl, getToken(), d)
}

func createRule(q string) []byte {
	v := fmt.Sprintf(`{"add": [{"value": "%s"}]}`, q)
	log.Output(2, v)
	return apiRequest("POST", rulesUrl, getToken(), v)
}

func saveTweet(b []byte) (string, error) {
	t, err := parseTweet(b)
	if err != nil {
		return "", err
	}

	n := fmt.Sprintf("data/%s.json", t.Id)
	ioutil.WriteFile(n, b, 0644)
	return n, nil
}

func stream() {
	req, err := http.NewRequest("GET", streamUrl, nil)
	crashIf(err)

	client := &http.Client{}
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Transfer-Encoding", "chunked")
	req.Header.Set("Authorization", "Bearer "+getToken())
	if a := os.Getenv("TWITTER_APP_NAME"); a != "" {
		req.Header.Set("User-agent", a)
	}

	res, err := client.Do(req)
	crashIf(err)
	defer res.Body.Close()

	s := bufio.NewScanner(res.Body)
	c := 0
	for s.Scan() {
		t := s.Bytes()
		if len(t) > 0 {
			c += 1
			go func(b []byte) {
				f, err := saveTweet(b)
				if err == nil {
					log.Output(2, fmt.Sprintf("[%d] %s saved", c, f))
				}
			}(t)
		}
	}
}

func main() {
	err := godotenv.Load()
	crashIf(err)

	app := &cli.App{
		Name: "Twitter Filtered Stream API client",
		Commands: []*cli.Command{
			{
				Name:  "rule",
				Usage: "Tools to mange the fiter rules",
				Subcommands: []*cli.Command{
					{
						Name:  "ls",
						Usage: "List the existing rules",
						Action: func(c *cli.Context) error {
							fmt.Println(string(getRules()))
							return nil
						},
					},
					{
						Name:  "rm",
						Usage: "Remove existing rules",
						Action: func(c *cli.Context) error {
							fmt.Println(string(deleteRules()))
							return nil
						},
					},
					{
						Name:  "new",
						Usage: "Create a rule",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "query",
								Usage: "Text for the rule",
							},
						},
						Action: func(c *cli.Context) error {
							fmt.Println(string(createRule(c.String("query"))))
							return nil
						},
					},
				},
			},
			{
				Name:  "stream",
				Usage: "Start streaming with current rules",
				Action: func(c *cli.Context) error {
					log.Output(2, "Listening to the stream…")
					stream()
					return nil
				},
			},
			{
				Name:  "token",
				Usage: "Show API bearer token",
				Action: func(c *cli.Context) error {
					fmt.Println(getToken())
					return nil
				},
			},
		},
	}
	app.Run(os.Args)

}
