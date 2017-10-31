/*
 * Copyright (c) 2017 WIFIPIX
 * Author: typdefxiaolu@gmail.com
 * Created Time: 2017/8/14
 *
 */

package utils

import (
    "math"
)

var PI float64 = 3.14159265358979324

type GPS struct {
    X_PI float64
}


func NewGPS() *GPS {
    return &GPS{X_PI:3.14159265358979324 * 3000.0 / 180.0,}
}

//WGS-84 to BD-09
func (gps *GPS) Wgs_bd( wgsLat float64, wgsLon  float64) map[string]float64 {
    coordinate := gps.Gcj_encrypt(wgsLat, wgsLon)
    return gps.Bd_encrypt(coordinate["lat"], coordinate["lon"])
}

//BD-09 to WGS-84
func (gps *GPS) Bd_wgs( bdLat float64, bdLon float64) map[string]float64 {
    coordinate := gps.Bd_decrypt(bdLat, bdLon)
    return gps.Gcj_decrypt_exact(coordinate["lat"], coordinate["lon"])
}

//WGS-84 to GCJ-02
func (gps *GPS) Gcj_encrypt( wgsLat float64, wgsLon float64) map[string]float64 {
    if gps.outOfChina(wgsLat, wgsLon) {
        return map[string]float64{"lat":wgsLat, "lon":wgsLon}
    }
    d := gps.delta(wgsLat, wgsLon)
    return map[string]float64{"lat": wgsLat + d["lat"],"lon": wgsLon + d["lon"]}
}

//GCJ-02 to WGS-84
func (gps *GPS) Gcj_decrypt( gcjLat float64, gcjLon float64) map[string]float64 {
    if gps.outOfChina(gcjLat, gcjLon) {
        return map[string]float64{"lat": gcjLat, "lon": gcjLon}
    }
    d := gps.delta(gcjLat, gcjLon)
    return map[string]float64{"lat": gcjLat - d["lat"], "lon": gcjLon - d["lon"]}
}

//GCJ-02 to WGS-84 exactly
func (gps *GPS) Gcj_decrypt_exact( gcjLat float64, gcjLon float64) map[string]float64 {
    initDelta := 0.01
    threshold := 0.000000001
    dLat := initDelta
    dLon := initDelta
    mLat := gcjLat - dLat
    mLon := gcjLon - dLon
    pLat := gcjLat + dLat
    pLon := gcjLon + dLon
    wgsLat := 0.
    wgsLon := 0.
    i := 0
    for {
        wgsLat = (mLat + pLat) / 2
        wgsLon = (mLon + pLon) / 2
        tmp := gps.Gcj_encrypt(wgsLat, wgsLon)
        dLat = tmp["lat"] - gcjLat
        dLon = tmp["lon"] - gcjLon
        if math.Abs(dLat) < threshold && math.Abs(dLon) < threshold {
            break
        }
        if dLat > 0 {
            pLat = wgsLat
        }else{
            mLat = wgsLat
        }
        if dLon > 0 {
            pLon = wgsLon
        }else{
            mLon = wgsLon
        }
        i += 1
        if i > 10000 {
            break
        }
    }
    return map[string]float64{"lat": wgsLat, "lon": wgsLon}
}
//GCJ-02 to BD-09
func (gps *GPS) Bd_encrypt( gcjLat float64, gcjLon float64) map[string]float64 {
    x := gcjLon
    y := gcjLat
    z := math.Sqrt(x * x + y * y) + 0.00002 * math.Sin(y * gps.X_PI)
    theta := math.Atan2(y, x) + 0.000003 * math.Cos(x * gps.X_PI)
    bdLon := z * math.Cos(theta) + 0.0065
    bdLat := z * math.Sin(theta) + 0.006
    return map[string]float64{"lat": bdLat,"lon": bdLon}
}

//BD-09 to GCJ-02
func (gps *GPS) Bd_decrypt( bdLat float64, bdLon float64) map[string]float64 {
    x := bdLon - 0.0065
    y := bdLat - 0.006
    z := math.Sqrt(x * x + y * y) - 0.00002 * math.Sin(y * gps.X_PI)
    theta := math.Atan2(y, x) - 0.000003 * math.Cos(x * gps.X_PI)
    gcjLon := z * math.Cos(theta)
    gcjLat := z * math.Sin(theta)
    return map[string]float64{"lat": gcjLat, "lon": gcjLon}
}
//WGS-84 to Web mercator
//mercatorLat.y  mercatorLon.x
func (gps *GPS) Mercator_encrypt( wgsLat float64, wgsLon float64) map[string]float64 {
    x := wgsLon * 20037508.34 / 180.
    y := math.Log(math.Tan((90. + wgsLat) * PI / 360.)) / (PI / 180.)
    y = y * 20037508.34 / 180.
    return map[string]float64{"lat": y, "lon": x}

}

// Web mercator to WGS-84
// mercatorLat.y mercatorLon.x
func (gps *GPS) Mercator_decrypt( mercatorLat float64, mercatorLon float64) map[string]float64 {
    x := mercatorLon / 20037508.34 * 180.
    y := mercatorLat / 20037508.34 * 180.
    y = 180 / PI * (2 * math.Atan(math.Exp(y * PI / 180.)) - PI / 2)
    return map[string]float64{"lat": y, "lon": x}
}

// two point"s distance
func (gps *GPS) Distance( latA float64, lonA float64, latB float64, lonB float64) float64 {
    earthR := 6371000.
    x := math.Cos(latA * PI / 180.) * math.Cos(latB * PI / 180.) * math.Cos((lonA - lonB) * PI / 180)
    y := math.Sin(latA * PI / 180.) * math.Sin(latB * PI / 180.)
    s := x + y
    if s > 1 { s = 1}
    if s < -1{ s = -1}
    alpha := math.Acos(s)
    distance := alpha * earthR
    return distance
}

func (gps *GPS) delta( lat float64, lon float64) map[string]float64 {

    // Krasovsky 1940
    //
    // a = 6378245.0, 1/f = 298.3
    // b = a * (1 - f)
    // ee = (a^2 - b^2) / a^2
    a := 6378245.0//  a: 卫星椭球坐标投影到平面地图坐标系的投影因子。
    ee := 0.00669342162296594323//  ee: 椭球的偏心率。
    dLat := gps.transformLat(lon - 105.0, lat - 35.0)
    dLon := gps.transformLon(lon - 105.0, lat - 35.0)
    radLat := lat / 180.0 * PI
    magic := math.Sin(radLat)
    magic = 1 - ee * magic * magic
    SqrtMagic := math.Sqrt(magic)
    dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * SqrtMagic) * PI)
    dLon = (dLon * 180.0) / (a / SqrtMagic * math.Cos(radLat) * PI)
    return map[string]float64{"lat": dLat, "lon": dLon}
}

func (gps *GPS) outOfChina( lat float64, lon float64) bool {
    if lon < 72.004 || lon > 137.8347 {
        return true
    }
    if lat < 0.8293 || lat > 55.8271 {
        return true
    }
    return false
}

func (gps *GPS) transformLat( x float64, y float64)  float64{
    ret := -100.0 + 2.0 * x + 3.0 * y + 0.2 * y * y + 0.1 * x * y + 0.2 * math.Sqrt(math.Abs(x))
    ret += (20.0 * math.Sin(6.0 * x * PI) + 20.0 * math.Sin(2.0 * x * PI)) * 2.0 / 3.0
    ret += (20.0 * math.Sin(y * PI) + 40.0 * math.Sin(y / 3.0 * PI)) * 2.0 / 3.0
    ret += (160.0 * math.Sin(y / 12.0 * PI) + 320 * math.Sin(y * PI / 30.0)) * 2.0 / 3.0
    return ret
}

func (gps *GPS) transformLon( x float64, y float64) float64{
    ret := 300.0 + x + 2.0 * y + 0.1 * x * x + 0.1 * x * y + 0.1 * math.Sqrt(math.Abs(x))
    ret += (20.0 * math.Sin(6.0 * x * PI) + 20.0 * math.Sin(2.0 * x * PI)) * 2.0 / 3.0
    ret += (20.0 * math.Sin(x * PI) + 40.0 * math.Sin(x / 3.0 * PI)) * 2.0 / 3.0
    ret += (150.0 * math.Sin(x / 12.0 * PI) + 300.0 * math.Sin(x / 30.0 * PI)) * 2.0 / 3.0
    return ret
}

// vim: set ts=8 sw=4 tw=0 :