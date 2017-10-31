# gisbeacon
Get the info of the Chinese administrative region by latitude and longitude.


## GIS Area Info API

### 1 请求地址

>http://localhost:8492/gis/area

### 2 调用方式：HTTP post

### 3 接口描述：

* 获取经纬度的行政区域归属信息

### 4 请求参数:



```json
{
    "type": "wgs84",
    "coordinates": [
        116.582808,
        40.084176
    ]
}
```

#### POST参数:
| 字段名称        | 字段说明 |   类型   |  必填  | 备注                          |
| ----------- | :--- | :----: | :--: | :-------------------------- |
| type        | 坐标类型 | string |  Y   | 支持类型 `gcj02`,`bd09`,`wgs84` |
| coordinates | 坐标点  | array  |  Y   | 注意经纬度顺序                     |



### 5 请求返回结果:

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


