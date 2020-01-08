package main

import (
    "os"
    "fmt"
    "errors"
    "strconv"
    "flag"
    "strings"
    "net/http"
    "os/user"
    "io/ioutil"
    "path/filepath"
    "sync"
    "github.com/tidwall/gjson"
    "github.com/spf13/viper"
    "github.com/mattn/go-runewidth"
    "golang.org/x/crypto/ssh/terminal"
)

const rcFile = "~/.aqirc"
const urlCityFeed = "https://api.waqi.info/feed/"
const urlSearch = "https://api.waqi.info/search/"
const rightJustifiedWidth = 50
const rightJustifiedWidthInZh = 30

type StringSlice []string

func (self *StringSlice) String() string {
    return fmt.Sprint(*self)
}

func (self *StringSlice) Set(value string) error {
    for _, ele := range strings.Split(value, ",") {
        *self = append(*self, ele)
    }
    return nil
}

var zhcn bool
var cities StringSlice
var keywords StringSlice
var token string
var basename string

func ExpandTildeToHomeDir(path string) (string, error) {
    if path[0] == '~' {
        cu, err := user.Current()
        if err != nil {
            return path, err
        }
        path = filepath.Join(cu.HomeDir, path[1:])
    }
    return path, nil
}

func IsInteger(s string) bool {
    if _, err := strconv.ParseInt(s, 10, 64); err == nil {
        return true
    }
    return false
}

func init() {
    basename = filepath.Base(os.Args[0])
    viper.SetDefault("zhcn", false)
    viper.SetConfigType("toml")
    configPath, _ := ExpandTildeToHomeDir(rcFile)
    viper.SetConfigFile(configPath)
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(*os.PathError); !ok {
            fmt.Fprintf(os.Stderr, "%s: Fatal error config file: %s\n", basename, err)
        }
    }
    zhcn = viper.GetBool("zhcn")
    cities = viper.GetStringSlice("cities")
    keywords = viper.GetStringSlice("keywords")
    token = viper.GetString("token")

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", basename)
        flag.PrintDefaults()
    }
    flag.BoolVar(&zhcn, "zhcn", zhcn, "中文显示")
    flag.BoolVar(&zhcn, "z", zhcn, "中文显示")
    var flagCities StringSlice
    flag.Var(&flagCities, "city", "Comma-separated list of cities.")
    flag.Var(&flagCities, "c", "Comma-separated list of cities.")
    var flagKeywords StringSlice
    flag.Var(&flagKeywords, "search", "Comma-separated list of keywords.")
    flag.Var(&flagKeywords, "s", "Comma-separated list of keywords.")
    flag.StringVar(&token, "token", token, "API token.")
    flag.StringVar(&token, "t", token, "API token.")
    flag.Parse()
    if len(flagCities) > 0 {
        cities = flagCities
    }
    if len(flagKeywords) > 0 {
        keywords = flagKeywords
    }
    if len(cities) == 0 && len(keywords) == 0 {
        cities = append(cities, "here")
    }
    if token == "" {
        fmt.Println("Must pass token")
        os.Exit(1)
    }
}

var errUnexpectedResponseFormat = errors.New("Unexpected response format")

var cityFeedResponseMappingTable = [][]string {
    { "data.city.name",     "Name of the monitoring station",   "监测站点"      },
    { "data.aqi",           "Real-time air quality",            "实时空气质量"  },
    { "data.iaqi.co.v",     "Carbon Monoxyde",                  "一氧化碳"      },
    { "data.iaqi.no2.v",    "Nitrogen Dioxide",                 "二氧化氮"      },
    { "data.iaqi.o3.v",     "Ozone",                            "臭氧"          },
    { "data.iaqi.pm10.v",   "PM10",                             "PM10"          },
    { "data.iaqi.pm25.v",   "PM2.5",                            "PM2.5"         },
    { "data.iaqi.so2.v",    "Sulphur Dioxide",                  "二氧化硫"      },
    { "data.iaqi.w.v",      "Wind",                             "气流"          },
    { "data.iaqi.t.v",      "Temperature",                      "气温"          },
    { "data.iaqi.r.v",      "Rain (precipitation)",             "降水"          },
    { "data.iaqi.h.v",      "Relative Humidity",                "相对湿度"      },
    { "data.iaqi.d.v",      "Dew",                              "露水"          },
    { "data.iaqi.p.v",      "Atmostpheric Pressure",            "气压"          },
    { "data.time.s",        "UTC time",                         "UTC时间"       },
}

var aqiColorChart = [][]int {
    { 0,    7  },
    { 50,   10  },
    { 100,  11  },
    { 150,  3   },
    { 200,  9   },
    { 300,  5   },
    { 500,  1   },
}

func colored(fg int, text string) string {
    if terminal.IsTerminal(int(os.Stdout.Fd())) {
        return fmt.Sprintf("\033[38;5;%dm%s\033[0m", fg, text)
    } else {
        return text
    }
}

func prettyCityFeed(parsedJson *gjson.Result) {
    aqi := parsedJson.Get("data.aqi").Float()
    i := 0
    for ; i < len(aqiColorChart) - 1; i++ {
        if aqi <= float64(aqiColorChart[i][0]) {
            break
        }
    }
    fg := aqiColorChart[i][1]
    for _, it := range cityFeedResponseMappingTable {
        if parsedJson.Get(it[0]).Exists() {
            if zhcn {
                fmt.Printf("%s: %s\n", runewidth.FillLeft(colored(fg, it[2]), rightJustifiedWidthInZh), colored(fg, parsedJson.Get(it[0]).String()))
            } else {
                fmt.Printf("%*s: %s\n", rightJustifiedWidth, colored(fg, it[1]), colored(fg, parsedJson.Get(it[0]).String()))
            }
        }
    }
    fmt.Println()
}

func apiCityFeed(city string) (*gjson.Result, error) {
    if IsInteger(city) {
        city = "@" + city
    }
    response, err := http.Get(urlCityFeed + city + "/?token=" + token)
    if err != nil {
		return nil, err
	}
    defer response.Body.Close()
    bytes, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return nil, err
    }
    parsedJson := gjson.Parse(string(bytes))
    if !parsedJson.Get("status").Exists() {
        return nil, errUnexpectedResponseFormat
    }
    if parsedJson.Get("status").Value().(string) != "ok" {
        return nil, errors.New(parsedJson.Get("data").Value().(string))
    }
    if !parsedJson.Get("data.city.name").Exists() || !parsedJson.Get("data.aqi").Exists() {
        return nil, errUnexpectedResponseFormat
    }
    return &parsedJson, nil
}

func apiSearch(keyword string) ([]*gjson.Result, []error) {
    var outsl []*gjson.Result
    var errsl []error
    response, err := http.Get(urlSearch + "/?token=" + token + "&keyword=" + keyword)
    if err != nil {
        errsl = append(errsl, err)
		return outsl, errsl
	}
    defer response.Body.Close()
    bytes, err := ioutil.ReadAll(response.Body)
    if err != nil {
        errsl = append(errsl, err)
        return outsl, errsl
    }
    parsedJson := gjson.Parse(string(bytes))
    if !parsedJson.Get("status").Exists() {
        errsl = append(errsl, errUnexpectedResponseFormat)
        return outsl, errsl
    }
    if parsedJson.Get("status").Value().(string) != "ok" {
        errsl = append(errsl, errors.New(parsedJson.Get("data").Value().(string)))
        return outsl, errsl
    }
    if !parsedJson.Get("data").Exists() {
        errsl = append(errsl, errUnexpectedResponseFormat)
        return outsl, errsl
    }
    for _, st := range parsedJson.Get("data").Array() {
        if out, err := apiCityFeed(st.Get("uid").String()); err != nil {
            errsl = append(errsl, err)
        } else {
            outsl = append(outsl, out)
        }
    }
    return outsl, errsl
}

func main() {
    var wg sync.WaitGroup
    chout := make(chan *gjson.Result)
    cherr := make(chan error)

    cityWorker := func(city string) {
        if out, err := apiCityFeed(city); err != nil {
            cherr <- err
        } else {
            chout <- out
        }
        wg.Done()
    }
    for _, city := range cities {
        wg.Add(1)
        go cityWorker(city)
    }

    searchWorker := func(keyword string) {
        outsl, errsl := apiSearch(keyword)
        for _, o := range outsl {
            chout <- o
        }
        for _, e := range errsl {
            cherr <- e
        }
        wg.Done()
    }
    for _, keyword := range keywords {
        wg.Add(1)
        go searchWorker(keyword)
    }

    done := make(chan struct{})
    go func() {
        wg.Wait()
        done <- struct{}{}
    }()
    for {
        select {
        case out := <-chout:
            prettyCityFeed(out)
        case err := <-cherr:
            fmt.Fprintf(os.Stderr, "%s: %s\n", basename, err)
        case <-done:
            return
        }
    }
}
