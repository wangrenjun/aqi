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
    "github.com/tidwall/gjson"
    "github.com/spf13/viper"
    "github.com/mattn/go-runewidth"
    "golang.org/x/crypto/ssh/terminal"
)

const rcFile = "~/.aqirc"
const urlCityFeed = "https://api.waqi.info/feed/"
const urlSearch = "https://api.waqi.info/search/"
const rightJustifiedWidth = 50

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
    viper.SetDefault("cities", "here")
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

func prettyCityFeed(parsedJson gjson.Result) {
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
                fmt.Printf("%s: %s\n", runewidth.FillLeft(colored(fg, it[2]), rightJustifiedWidth), colored(fg, parsedJson.Get(it[0]).String()))
            } else {
                fmt.Printf("%*s: %s\n", rightJustifiedWidth, colored(fg, it[1]), colored(fg, parsedJson.Get(it[0]).String()))
            }
        }
    }
    fmt.Println()
}

func apiCityFeed(city string) error {
    if IsInteger(city) {
        city = "@" + city
    }
    response, err := http.Get(urlCityFeed + city + "/?token=" + token)
    if err != nil {
		return err
	}
    defer response.Body.Close()
    bytes, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return err
    }
    parsedJson := gjson.Parse(string(bytes))
    if !parsedJson.Get("status").Exists() {
        return errUnexpectedResponseFormat
    }
    if parsedJson.Get("status").Value().(string) != "ok" {
        return errors.New(parsedJson.Get("data").Value().(string))
    }
    if !parsedJson.Get("data.city.name").Exists() || !parsedJson.Get("data.aqi").Exists() {
        return errUnexpectedResponseFormat
    }
    prettyCityFeed(parsedJson)
    return nil
}

func apiSearch(keyword string) error {
    response, err := http.Get(urlSearch + "/?token=" + token + "&keyword=" + keyword)
    if err != nil {
		return err
	}
    defer response.Body.Close()
    bytes, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return err
    }
    parsedJson := gjson.Parse(string(bytes))
    if !parsedJson.Get("status").Exists() {
        return errUnexpectedResponseFormat
    }
    if parsedJson.Get("status").Value().(string) != "ok" {
        return errors.New(parsedJson.Get("data").Value().(string))
    }
    if !parsedJson.Get("data").Exists() {
        return errUnexpectedResponseFormat
    }
    for _, st := range parsedJson.Get("data").Array() {
        if err := apiCityFeed(st.Get("uid").String()); err != nil {
            fmt.Fprintf(os.Stderr, "%s: %s\n", basename, err)
        }
    }
    return nil
}

func main() {
    for _, c := range cities {
        if err := apiCityFeed(c); err != nil {
            fmt.Fprintf(os.Stderr, "%s: %s\n", basename, err)
        }
    }
    for _, k := range keywords {
        if err := apiSearch(k); err != nil {
            fmt.Fprintf(os.Stderr, "%s: %s\n", basename, err)
        }
    }
}
