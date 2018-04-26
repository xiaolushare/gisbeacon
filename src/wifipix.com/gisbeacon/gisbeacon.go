/*
 * Copyright (c) 2017 WIFIPIX
 * Author: typdefxiaolu@gmail.com
 * Created Time: 2017/9/29
 *
 */

package main

import (
    "wifipix.com/gisbeacon/s2"
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
    "io/ioutil"
    "github.com/xeipuuv/gojsonschema"
    "encoding/json"
    "flag"
    "runtime"
    "wifipix.com/gisbeacon/wp"
    "wifipix.com/gisbeacon/utils"
)

const CoordinatesTypePattern = `"^([Ww][Gg][Ss]84)$|^([Gg][Cc][Jj]02)$|^([Bb][Dd]09)$"`

const CoordinatesSchema = `
                           "coordinates": {
                                 "type": "array",
                                 "minItems": 2,
                                 "maxItems": 2,
                                 "items":[
                                        {
                                          "type": "number",
                                          "minimum": -180,
                                          "maximum": 180,
                                          "exclusiveMaximum": true,
                                          "exclusiveMinimum": true
                                        },
                                        {
                                          "type": "number",
                                          "minimum": -90,
                                          "maximum": 90,
                                          "exclusiveMaximum": true,
                                          "exclusiveMinimum": true
                                        }
                                 ]
                           } `

const AreaLocationRequestSchema = `{
                        "type": "object",
                        "properties": {
                               "type": {
                                    "type": "string",
                                    "pattern": ` + CoordinatesTypePattern + `
                                    },
                              ` + CoordinatesSchema + `
                            },
                        "required":["type", "coordinates"]
                        }`

const CoordinatesConvertRequestSchema = `{
                        "type": "object",
                        "properties": {
                               "type": {
                                    "type": "string",
                                    "pattern": ` + CoordinatesTypePattern + `
                                    },
                              ` + CoordinatesSchema + `,
                               "convert": {
                                    "type": "string",
                                    "pattern": ` + CoordinatesTypePattern + `
                                }
                            },
                        "required":["type", "convert", "coordinates"]
                        }`

type CoordinatesConvertRequest struct {
    s2.Loc
    Convert string `json:"convert"`
}

const CoordinatesGeohashRequestSchema = `{
                        "type": "object",
                        "properties": {
                            ` + CoordinatesSchema + `,
                            "precision": {
                                "type": "number",
                                "minimum": 1,
                                "maximum": 12,
                                "exclusiveMaximum": true,
                                "exclusiveMinimum": true
                            }
                        },
                        "required": ["coordinates"]
                        }`

type CoordinatesGeohashRequest struct {
    Coordinates []float64 `json:"coordinates"`
    Precision   int       `json:"precision"`
}

const CoordinatesPolygonRelationRequestSchema = `{
                        "type": "object",
                        "properties": {
                              ` + CoordinatesSchema + `,
                               "polygon": {
                                    "type": "array",
                                    "minItems": 4,
                                    "maxItems": 1000,
                                    "items":[
                                        {
                                             "type": "array",
                                             "minItems": 2,
                                             "maxItems": 2,
                                             "items":[
                                                    {
                                                      "type": "number",
                                                      "minimum": -180,
                                                      "maximum": 180,
                                                      "exclusiveMaximum": true,
                                                      "exclusiveMinimum": true
                                                    },
                                                    {
                                                      "type": "number",
                                                      "minimum": -90,
                                                      "maximum": 90,
                                                      "exclusiveMaximum": true,
                                                      "exclusiveMinimum": true
                                                    }
                                             ]
                                        }
                                    ]
                               }
                            },
                        "required": ["polygon", "coordinates"]
                        }`

type CoordinatesPolygonRelationRequest struct {
    Coordinates [2]float64 `json:"coordinates"`
    wp.Polygon
}

var area_path = flag.String("area", "./data/area.json", "area content json file path")
var geohash_path = flag.String("geohash", "./data/geohash5adcodemap.json", "geohash 5 adcode map json file path")
var port = flag.String("port", "8492", "server listen port")

func main() {
    runtime.GOMAXPROCS(0)
    flag.Parse()
    var areaLocationRequestSchema = gojsonschema.NewStringLoader(AreaLocationRequestSchema)
    var coordinatesConvertRequestSchema = gojsonschema.NewStringLoader(CoordinatesConvertRequestSchema)
    var coordinatesGeohashRequestSchema = gojsonschema.NewStringLoader(CoordinatesGeohashRequestSchema)
    var coordinatesPolygonRelationRequestSchema = gojsonschema.NewStringLoader(CoordinatesPolygonRelationRequestSchema)
    area := s2.NewGisArea(*area_path, *geohash_path)
    gps := utils.NewGPS()

    var setContextHearder = func(c *gin.Context) {
        c.Writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
        c.Writer.Header().Set("Server", "wifipix location server/1.0")
    }

    var requestJsonSchemaValidate = func(schema gojsonschema.JSONLoader, c *gin.Context) ([]byte, bool) {
        body := c.Request.Body
        reqBody, _ := ioutil.ReadAll(body)
        loder := gojsonschema.NewStringLoader(string(reqBody))
        r, _ := gojsonschema.Validate(schema, loder)
        if len(r.Errors()) > 0 {
            //fmt.Println("err:%v", r.Errors())
            c.JSON(http.StatusBadRequest, gin.H{
                "error": map[string]interface{}{
                    "code":    -1,
                    "message": fmt.Sprintf("parameter request error %v", r.Errors()),
                },
            })
            return nil, false
        }
        return reqBody, true
    }

    r := gin.Default()
    // 经纬度坐标行政归属地
    r.POST("/gis/area", func(c *gin.Context) {
        setContextHearder(c)
        var loc s2.Loc
        if reqBody, validate := requestJsonSchemaValidate(areaLocationRequestSchema, c); !validate {
            return
        } else {
            json.Unmarshal(reqBody, &loc)
        }
        result := area.GetContainArea(loc)

        if result != nil {
            area := map[string]interface{}{
                "code": result.Adcode,
                "name": result.Name,
                //"district": result.District,
                //"city": result.City,
                //"province": result.Province,
                //"parent_id": result.ParentId,
                //"loc": result.Loc,
            }

            if result.District != "" {
                area["district"] = result.District
            }
            if result.City != "" {
                area["city"] = result.City
            }
            if result.Province != "" {
                area["province"] = result.Province
            }
            c.JSON(http.StatusOK, gin.H{
                "area": area,
                "error": map[string]interface{}{
                    "code":    0,
                    "message": "success",
                },
            })
        } else {
            c.JSON(http.StatusNotFound, gin.H{
                "error": map[string]interface{}{
                    "code":    -2,
                    "message": "area not found may be out of china",
                },
            })
        }
    })
    // 经纬度坐标系转换
    r.POST("/gis/convert", func(c *gin.Context) {
        setContextHearder(c)
        var coordinatesConvertRequest CoordinatesConvertRequest
        var resCoordinates = &s2.Loc{Type: coordinatesConvertRequest.Convert}
        if reqBody, validate := requestJsonSchemaValidate(coordinatesConvertRequestSchema, c); !validate {
            return
        } else {
            json.Unmarshal(reqBody, &coordinatesConvertRequest)
        }
        switch coordinatesConvertRequest.Type {
        case "wgs84":
            switch coordinatesConvertRequest.Convert {
            case "gcj02":
                coords := gps.Gcj_encrypt(coordinatesConvertRequest.Coordinates[1], coordinatesConvertRequest.Coordinates[0])
                resCoordinates.Coordinates = []float64{coords["lon"], coords["lat"]}
            case "bd09":
                coords := gps.Wgs_bd(coordinatesConvertRequest.Coordinates[1], coordinatesConvertRequest.Coordinates[0])
                resCoordinates.Coordinates = []float64{coords["lon"], coords["lat"]}
            default:
                resCoordinates.Coordinates = coordinatesConvertRequest.Coordinates
            }
        case "gcj02":
            switch coordinatesConvertRequest.Convert {
            case "wgs84":
                coords := gps.Gcj_decrypt_exact(coordinatesConvertRequest.Coordinates[1], coordinatesConvertRequest.Coordinates[0])
                resCoordinates.Coordinates = []float64{coords["lon"], coords["lat"]}
            case "bd09":
                coords := gps.Bd_encrypt(coordinatesConvertRequest.Coordinates[1], coordinatesConvertRequest.Coordinates[0])
                resCoordinates.Coordinates = []float64{coords["lon"], coords["lat"]}
            default:
                resCoordinates.Coordinates = coordinatesConvertRequest.Coordinates
            }
        case "bd09":
            switch coordinatesConvertRequest.Convert {
            case "wgs84":
                coords := gps.Bd_wgs(coordinatesConvertRequest.Coordinates[1], coordinatesConvertRequest.Coordinates[0])
                resCoordinates.Coordinates = []float64{coords["lon"], coords["lat"]}
            case "gcj02":
                coords := gps.Bd_decrypt(coordinatesConvertRequest.Coordinates[1], coordinatesConvertRequest.Coordinates[0])
                resCoordinates.Coordinates = []float64{coords["lon"], coords["lat"]}
            default:
                resCoordinates.Coordinates = coordinatesConvertRequest.Coordinates
            }
        }
        c.JSON(http.StatusOK, gin.H{
            "result": resCoordinates,
            "error": map[string]interface{}{
                "code":    0,
                "message": "success",
            },
        })
    })
    // 经纬度坐标geohash值以及边界值
    r.POST("/gis/geohash", func(c *gin.Context) {
        setContextHearder(c)
        var coordinatesGeohashRequest CoordinatesGeohashRequest
        if reqBody, validate := requestJsonSchemaValidate(coordinatesGeohashRequestSchema, c); !validate {
            return
        } else {
            json.Unmarshal(reqBody, &coordinatesGeohashRequest)
            if coordinatesGeohashRequest.Precision == 0 {
                coordinatesGeohashRequest.Precision = 12
            }
        }
        hash, box := utils.Encode(coordinatesGeohashRequest.Coordinates[1], coordinatesGeohashRequest.Coordinates[0], coordinatesGeohashRequest.Precision)
        c.JSON(http.StatusOK, gin.H{
            "result": map[string]interface{}{
                "geohash":   hash,
                "precision": coordinatesGeohashRequest.Precision,
                "box":       box,
            },
            "error": map[string]interface{}{
                "code":    0,
                "message": "success",
            },
        })
    })

    // 经纬度坐标是否在围栏内判定
    r.POST("/gis/polygon/contains", func(c *gin.Context) {
        setContextHearder(c)
        var coordinatesPolygonRelationRequest CoordinatesPolygonRelationRequest
        if reqBody, validate := requestJsonSchemaValidate(coordinatesPolygonRelationRequestSchema, c); !validate {
            return
        } else {
            json.Unmarshal(reqBody, &coordinatesPolygonRelationRequest)
        }
        polygon := wp.NewPolygon(coordinatesPolygonRelationRequest.Path)
        var result = polygon.Contains(coordinatesPolygonRelationRequest.Coordinates)
        c.JSON(http.StatusOK, gin.H{
            "result": map[string]bool {"contains": result},
            "error": map[string]interface{}{
                "code":    0,
                "message": "success",
            },
        })
    })

    r.Run(":" + *port)
}

// 116.590479,40.083564 朝阳
// 116.582808,40.084176 朝阳
// 116.753661,40.152315 顺义

// vim: set ts=8 sw=4 tw=0 :
