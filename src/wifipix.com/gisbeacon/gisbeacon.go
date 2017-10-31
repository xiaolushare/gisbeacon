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
)

const CoordinatesSchema = `{
                        "type": "object",
                        "properties": {
                               "type": {
                                    "type": "string",
                                    "pattern": "^([Ww][Gg][Ss]84)$|^([Gg][Cc][Jj]02)$|^([Bb][Dd]09)$"
                                    },
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
                               }
                            },
                        "required":["type", "coordinates"]
                        }`


var area_path = flag.String("area", "./data/area.json", "area content json file path")
var geohash_path = flag.String("geohash", "./data/geohash5adcodemap.json", "geohash 5 adcode map json file path")
var port = flag.String("port", "8492", "server listen port")

func main() {
    runtime.GOMAXPROCS(0)
    flag.Parse()
    var schema = gojsonschema.NewStringLoader(CoordinatesSchema)
    area := s2.NewGisArea(*area_path, *geohash_path)
    r := gin.Default()
    r.POST("/gis/area", func(c *gin.Context) {
        c.Writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
        c.Writer.Header().Set("Server", "wifipix location server/1.0")
        body := c.Request.Body
        reqBody, _ := ioutil.ReadAll(body)
        loder := gojsonschema.NewStringLoader(string(reqBody))
        r, _ := gojsonschema.Validate(schema, loder)
        if len(r.Errors()) > 0 {
            fmt.Errorf("err:%v", r.Errors())
            c.JSON(http.StatusBadRequest, gin.H{
                "error": map[string ] interface{} {
                    "code":    -1,
                    "message": "parameter request error",
                },
            })
            return
        }

        var loc s2.Loc

        json.Unmarshal(reqBody, &loc)

        result := area.GetContainArea(loc)

        if result != nil {
            area := map[string] interface{}{
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
                "error": map[string ] interface{} {
                "code":    0,
                "message": "ok",
                },
            })
        }else {
            c.JSON(http.StatusNotFound, gin.H{
                "error": map[string ] interface{} {
                    "code":    -2,
                    "message": "area not found may be out of china",
                },
            })
        }
    })
    r.Run(":"+*port)
}

// 116.590479,40.083564 朝阳
// 116.582808,40.084176 朝阳
// 116.753661,40.152315 顺义

// vim: set ts=8 sw=4 tw=0 :