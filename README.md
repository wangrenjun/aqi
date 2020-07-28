# aqi

一个命令行查询实时空气质量指数的工具，使用Go开发，数据来自 [Real-time Air Quality Index](http://aqicn.org/)；

![screenshot1](https://user-images.githubusercontent.com/1911620/70865707-6c902780-1f9b-11ea-963a-67a9d00f2aa9.png)

![screenshot2](https://user-images.githubusercontent.com/1911620/70865708-6c902780-1f9b-11ea-84b0-04168745ad2f.png)

安装：
-------

```sh
$ go get -u github.com/wangrenjun/aqi
```

墙内用户（Go 1.13及以上）：

```sh
GO111MODULE=on GOPROXY=https://goproxy.cn go get -u github.com/wangrenjun/aqi
```

快速使用：
---------

查询本地天气（IP地址定位）：
```shell
$ ${GOPATH}/bin/aqi -t <TOKEN>
```

查询某城市的空气指数：
```shell
$ ${GOPATH}/bin/aqi -t <TOKEN> -c beijing
```

查询多个城市的空气指数：
```shell
$ ${GOPATH}/bin/aqi -t <TOKEN> -c beijing,shanghai -c guangzhou,shenzhen,hangzhou
```

查询某城市的所有监测站点的空气指数：
```shell
$ ${GOPATH}/bin/aqi -t <TOKEN> -s beijing
```

查询多个城市的所有监测站点的空气指数：
```shell
$ ${GOPATH}/bin/aqi -t <TOKEN> -s beijing,shanghai,guangzhou
```

以中文显示查询结果：
```shell
$ ${GOPATH}/bin/aqi -t <TOKEN> -c beijing,shanghai -c guangzhou,shenzhen,hangzhou -z
```
```shell
$ ${GOPATH}/bin/aqi -t <TOKEN> -s beijing,shanghai,guangzhou -z
```

命令选项：
-------

```shell
Usage of aqi:
    -c value
            Comma-separated list of cities.
    -city value
            Comma-separated list of cities.
    -s value
            Comma-separated list of keywords.
    -search value
            Comma-separated list of keywords.
    -t string
            API token.
    -token string
            API token.
    -z      中文显示
    -zhcn
            中文显示
```

关于Token的申请：
------------------
因为数据来自 [Real-time Air Quality Index](http://aqicn.org/)，所以使用前需要到[这里](https://aqicn.org/data-platform/token/#/)申请Token。

配置文件：
----------
配置文件的默认路径是 $HOME/.aqirc，文件类型为[Toml](https://github.com/toml-lang/toml)，配置好了之后就不需要每次输入参数，可配置的选项有：
 * `zhcn`           bool值，是否以中文显示；
 * `cities`         数组值，需要查询的城市名称（拼音）；
 * `keywords`       数组值，需要查询监测站点的城市名称（拼音）；
 * `token`          string值，就是Token；

一个配置例子：
```shell
zhcn = false
cities = [ "beijing", "shanghai", "dalian" ]
keywords = [ "beijing", "guangzhou" ]
token = "demo"
```

结语：
-----
因为服务在墙外，所以在未使用科学上网的情况下，偶尔会不稳定，但是总体上还是能够使用的。
