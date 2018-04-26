##GIS Beacon

> GIS 常用的一些工具类服务接口封装

- *经纬度坐标获取行政区域归属地信息*
- *经纬度坐标三种常用坐标系互转*
- *经纬度坐标GEOhash值以及其边界经纬度*
- *经纬度坐标是否在围栏内判定*


###GIS TOOLS API

####1 经纬度坐标获取行政区域归属地信息

#####使用说明

>请求方式：POST
>
>请求头部：Content-Type： application/json
>
>请求URL ：http://localhost:port/gis/area
>
>是否允许跨域：否

#####请求参数

> **request body example**

```json
{
    "type": "wgs84",
    "coordinates": [
        116.582808,
        40.084176
    ]
}
```

> **body**

| 字段名称        | 字段说明 |   类型   |  必填  | 备注                          |
| ----------- | :--- | :----: | :--: | :-------------------------- |
| type        | 坐标类型 | string |  Y   | 支持类型 `gcj02`,`bd09`,`wgs84` |
| coordinates | 坐标点  | array  |  Y   | 注意经纬度顺序 【经度，维度】                     |

>**response body example**

```json
{
  "area":{
    "code":110113,
    "name":"顺义区",
    "district":"顺义区",
    "city":"北京市",
    "province":"北京市"
  },
  "error":{
    "code":0,
    "message":"ok"
  }
}
```

| 字段名称        | 字段说明 |   类型   |  必填  | 备注                          |
| ----------- | :--- | :----: | :--: | :-------------------------- |
| code    | 行政区域编码 | number |  Y   | 参考http://xzqh.mca.gov.cn/map |
| name | 所属区域名称 | string |  Y   | 最小级别到区县级别          |
| district | 区县 | string | N | 无 |
| city | 市 | string | N | 无 |
| province | 省 | string | N | 无 |


#### 2 经纬度坐标三种常用坐标系互转

##### 使用说明

> 请求方式：POST
>
> 请求头部：Content-Type： application/json
>
> 请求URL ：http://localhost:port/gis/convert
>
> 是否允许跨域：否

##### 请求参数

> **request body example**

```json
{
    "type": "wgs84",
    "coordinates": [
        116.582808,
        40.084176
    ],
    "convert": "bd09"
}
```

> **body**

| 字段名称    | 字段说明 |  类型  | 必填 | 备注                            |
| ----------- | :------- | :----: | :--: | :------------------------------ |
| type        | 坐标类型 | string |  Y   | 支持类型 `gcj02`,`bd09`,`wgs84` |
| coordinates | 坐标点   | array  |  Y   | 注意经纬度顺序 【经度，维度】   |
| convert     | 转换类型 | string |  Y   | 支持类型 `gcj02`,`bd09`,`wgs84  |

> **response body example**

```json
{
    "error": {
        "code": 0,
        "message": "success"
    },
    "result": {
        "type": "bd09",
        "coordinates": [
            116.5952308827249,
            40.09091477235431
        ]
    }
}
```



#### 3 经纬度坐标GEOhash值以及其边界经纬度

##### 使用说明

> 请求方式：POST
>
> 请求头部：Content-Type： application/json
>
> 请求URL ：http://localhost:port/gis/geohash
>
> 是否允许跨域：否

##### 请求参数

> **request body example**

```json
{
    "precision": 12,
    "coordinates": [
        116.582808,
        40.084176
    ]
}
```

> **body**

| 字段名称    | 字段说明 |  类型  | 必填 | 备注                          |
| ----------- | :------- | :----: | :--: | :---------------------------- |
| precision   | 精度级别 | number |  N   | 范围（1~12）默认12            |
| coordinates | 坐标点   | array  |  Y   | 注意经纬度顺序 【经度，维度】 |

> **response body example**

```json
{
    "error": {
        "code": 0,
        "message": "success"
    },
    "result": {
        "box": {
            "minLat": 40.084175895899534,
            "maxLat": 40.0841760635376,
            "minLng": 116.58280793577433,
            "maxLng": 116.58280827105045
        },
        "geohash": "wx4uhcj50rfr",
        "precision": 12
    }
}
```

>body

| 字段名称 | 字段说明  |  类型  | 备注                                      |
| -------- | :-------- | :----: | :---------------------------------------- |
| box      | 边界值    | object | minLat,maxLat(纬度)   minLng,maxLng(经度) |
| geohash  | geohash值 | string | 参考https://en.wikipedia.org/wiki/Geohash |





#### 4 经纬度坐标是否在围栏内判定

##### 使用说明

> 请求方式：POST
>
> 请求头部：Content-Type： application/json
>
> 请求URL ：http://localhost:port/gis/polygon/contains
>
> 是否允许跨域：否

##### 请求参数

> **request body example**

```json
{
    "coordinates": [
        116.582808,
        40.084176
    ],
  "polygon": [
  	  [116.58280793577433, 40.084175895899534], [116.58280793577433, 40.084175895899534], 
      [116.58280793577433, 40.084175895899534], [116.58280793577433, 40.084175895899534]
  ]
}
```

> **body**

| 字段名称    | 字段说明   |     类型     | 必填 | 备注                                   |
| ----------- | :--------- | :----------: | :--: | :------------------------------------- |
| coordinates | 坐标点     |    array     |  Y   | 注意经纬度顺序 【经度，维度】          |
| polygon     | 坐标点集合 | array[array] |  Y   | 围栏为闭合2D平面，顺序为起始点到结束点 |

> **response body example**

```json
{
    "error": {
        "code": 0,
        "message": "success"
    },
    "result": {
        "contains": false
    }
}
```

| 字段名称 | 字段说明            | 类型 | 备注                                |
| -------- | :------------------ | :--: | :---------------------------------- |
| contains | 包含（true\|false） | bool | 包含（含边界）为true，不包含为false |


