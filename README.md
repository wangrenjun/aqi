# aqi

一个命令行查询实时空气质量指数的工具，使用Go开发，数据来自 [Real-time Air Quality Index](http://aqicn.org/)；

![screenshot1](https://user-images.githubusercontent.com/1911620/70865707-6c902780-1f9b-11ea-963a-67a9d00f2aa9.png)

![screenshot2](https://user-images.githubusercontent.com/1911620/70865708-6c902780-1f9b-11ea-84b0-04168745ad2f.png)

快速入门：
---------

查询本地天气（根据本地IP地址）：
```shell
go run ./aqi.go -t <TOKEN>
```

查询某城市的空气指数：
```shell
go run ./aqi.go -t <TOKEN> -c beijing
```

查询多个城市的空气指数：
```shell
go run ./aqi.go -t <TOKEN> -c beijing,shanghai -c guangzhou,shenzhen,hangzhou
```

查询某城市的所有监测站点的空气指数：
```shell
go run ./aqi.go -t <TOKEN> -s beijing
```

查询多个城市的所有监测站点的空气指数：
```shell
go run ./aqi.go -t <TOKEN> -s beijing,shanghai,guangzhou
```

以中文显示上面的查询：
```shell
go run ./aqi.go -t <TOKEN> -s beijing,shanghai,guangzhou -z
```

Usage：
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
在[这里](https://aqicn.org/data-platform/token/#/)申请。

配置文件：
----------
配置文件的默认路径是 $HOME/.aqirc，文件类型为Toml，配置以后就不需要每次录入参数了，可配置的选项有：
 * `zhcn`           bool值，是否以中文显示；
 * `cities`         数组值，需要查询的城市名称（拼音）；
 * `keywords`       数组值，需要查询监测站点的城市名称（拼音）；
 * `token`          string值，就是TOKEN；

一个例子：
```shell
zhcn = false
cities = [ "beijing", "shanghai", "dalian" ]
keywords = [ "beijing", "guangzhou" ]
token = "demo"
```

结语：
-----
因为服务在墙外，所以在未使用科学上网的情况下，偶尔会不稳定，但是总体上还是能够使用的。
