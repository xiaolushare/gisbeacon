/*
 * Copyright (c) 2017 WIFIPIX
 * Author: typdefxiaolu@gmail.com
 * Created Time: 2017/9/30
 *
 */

package s2

import (
    "fmt"
    "encoding/json"
    "github.com/golang/geo/s2"
    "os"
    "io/ioutil"
    "io"
    "bufio"
    "wifipix.com/gisbeacon/utils"
    "strings"
)

type Loc struct {
    Type        string     `json:"type"`
    Coordinates []float64  `json:"coordinates"`
}

type Polygon struct {
    Type        string                 `json:"type"`
    Coordinates [][][][]float64   `json:"coordinates"`
}

type Area struct {
    Province string     `json:"province"`
    City     string     `json:"city"`
    District string     `json:"district"`
    Name     string     `json:"name"`
    Adcode   int        `json:"adcode"`
    ParentId int        `json:"parent_id"`
    Loc      Loc        `json:"loc"`
    Polygon  Polygon    `json:"polygon"`
}

type GisArea struct {
    GPS         *utils.GPS
    Area map[int] *Area
    AreaPolygon map[int] [][]*s2.Polygon
    GeoHash5Area map[string] []int
}

func NewGisArea(areafile, geomapfile string) *GisArea {
    gisArea := &GisArea {
        GPS: utils.NewGPS(),
        Area: make(map[int] *Area),
        AreaPolygon: make(map[int] [][]*s2.Polygon),
        GeoHash5Area: make(map[string] []int),
    }
    //TODO read area json file
    f, err := os.Open(areafile)
    defer f.Close()
    if err != nil {
        fmt.Fprintln(os.Stderr, "open area json file error %v", err)
        panic(err)
    }
    buf := bufio.NewReader(f)
    for {
        line, err := buf.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                break
            }
            fmt.Fprintln(os.Stderr, "read area json file error %v", err)
            panic(err)
        }
        var area Area
        json.Unmarshal([]byte(line), &area)
        gisArea.Area[area.Adcode] = &area
        gisArea.AreaPolygon[area.Adcode] = make([][]*s2.Polygon, 0)
        // create s2 polygon
        for _, level1 := range area.Polygon.Coordinates {
            polygons := make([]*s2.Polygon, 0)
            for _, level2 := range level1 {
                poits := make([]s2.Point, 0)
                for _, loc := range level2 {
                    point := s2.PointFromLatLng(s2.LatLngFromDegrees(loc[1], loc[0]))
                    poits = append(poits, point)
                }
                loop := s2.LoopFromPoints(poits)
                polygon := s2.PolygonFromLoops([]*s2.Loop{loop})
                polygons = append(polygons, polygon)
            }
            gisArea.AreaPolygon[area.Adcode] = append(gisArea.AreaPolygon[area.Adcode], polygons)
        }
    }

    //TODO read geomap json file
    //f, err = os.Open(geomapfile)
    content, err := ioutil.ReadFile(geomapfile)
    defer f.Close()
    if err != nil {
        fmt.Fprintln(os.Stderr, "open area json file error %v", err)
        panic(err)
    }
    json.Unmarshal([]byte(content), &gisArea.GeoHash5Area)
    return gisArea
}

func (gis *GisArea) GetContainArea(reqLoc Loc) *Area {
    switch strings.ToLower(reqLoc.Type) {
    case "bd09":
        wloc := gis.GPS.Bd_wgs(reqLoc.Coordinates[1], reqLoc.Coordinates[0])
        reqLoc.Coordinates = []float64 {wloc["lon"], wloc["lat"], }
    case "gcj02":
        wloc := gis.GPS.Gcj_decrypt_exact(reqLoc.Coordinates[1], reqLoc.Coordinates[0])
        reqLoc.Coordinates = []float64 {wloc["lon"], wloc["lat"], }
    }
    //fmt.Println(reqLoc)
    geohash5, _ := utils.Encode(reqLoc.Coordinates[1], reqLoc.Coordinates[0], 5)
    //fmt.Println(geohash5)
    if _, ok := gis.GeoHash5Area[geohash5]; ok {
        //fmt.Println(gis.GeoHash5Area[geohash5])
        for _, adcode := range gis.GeoHash5Area[geohash5] {
            for _, polygons := range gis.AreaPolygon[adcode] {
                flag := false
                //fmt.Println(adcode, len(polygons))
                for i, polygon := range polygons {
                    if i == 0 && ! polygon.ContainsPoint(s2.PointFromLatLng(s2.LatLngFromDegrees(reqLoc.Coordinates[1], reqLoc.Coordinates[0]))) {
                        //fmt.Println(gis.Area[adcode])
                        flag = true
                    }else if !polygon.ContainsPoint(s2.PointFromLatLng(s2.LatLngFromDegrees(reqLoc.Coordinates[1], reqLoc.Coordinates[0]))) {
                        flag = false
                        //fmt.Println(gis.Area[adcode])
                        break
                    }
                }
                if flag {
                    return gis.Area[adcode]
                }
            }
        }

        // not area so just province
        for _, adcode := range gis.GeoHash5Area[geohash5] {
            if _, ok := gis.Area[gis.Area[adcode].ParentId]; ok {
                for _, polygons := range gis.AreaPolygon[gis.Area[adcode].ParentId] {
                    flag := false
                    for i, polygon := range polygons {
                        if i == 0 && ! polygon.ContainsPoint(s2.PointFromLatLng(s2.LatLngFromDegrees(reqLoc.Coordinates[1], reqLoc.Coordinates[0]))) {
                            //fmt.Println(gis.Area[adcode])
                            flag = true
                        }else if !polygon.ContainsPoint(s2.PointFromLatLng(s2.LatLngFromDegrees(reqLoc.Coordinates[1], reqLoc.Coordinates[0]))) {
                            flag = false
                            //fmt.Println(gis.Area[adcode])
                            break
                        }
                    }
                    if flag {
                        return gis.Area[gis.Area[adcode].ParentId]
                    }
                }
            }
        }
    }

    return nil
}

// vim: set ts=8 sw=4 tw=0 :