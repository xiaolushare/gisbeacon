/*
 * Copyright (c) 2017 WIFIPIX
 * Author: typdefxiaolu@gmail.com
 * Created Time: 2017/8/14
 *
 */

package utils

import (
    "bytes"
    "strings"
)

const (
    BASE32                = "0123456789bcdefghjkmnpqrstuvwxyz"
    MAX_LATITUDE  float64 = 90
    MIN_LATITUDE  float64 = -90
    MAX_LONGITUDE float64 = 180
    MIN_LONGITUDE float64 = -180
)

var (
    bits   = []int{16, 8, 4, 2, 1}
    base32 = []byte(BASE32)
)

// geohash的精度与其长度成正比
// 每个点的geohash值实际上代表了一个区域，这个区域的大小与geohash的精度成反比
// 坐标点的格式为（纬度，经度）
// 将这个区域用一个矩形表示
type Box struct {
    MinLat float64 `json:"minLat"` // 维度
    MaxLat float64 `json:"maxLat"`
    MinLng float64 `json:"minLng"`
    MaxLng float64 `json:"maxLng"` // 经度
}

func (this *Box) Width() float64 {
    return this.MaxLng - this.MinLng
}

func (this *Box) Height() float64 {
    return this.MaxLat - this.MinLat
}

// geohash精度的设定参考 http://en.wikipedia.org/wiki/Geohash
// geohash length	lat   bits	lng bits	lat error	lng error	km error
// 1				2			3			±23				±23				±2500
// 2				5			5			± 2.8			± 5.6			±630
// 3				7			8			± 0.70		    ± 0.7			±78
// 4				10		    10		    ± 0.087		    ± 0.18		    ±20
// 5				12		    13		    ± 0.022		    ± 0.022		    ±2.4
// 6				15		    15		    ± 0.0027	    ± 0.0055	    ±0.61
// 7				17		    18		    ±0.00068	    ±0.00068	    ±0.076
// 8				20		    20		    ±0.000085	    ±0.00017	    ±0.019

// 输入值：纬度，经度，精度(geohash的长度)
// 返回geohash, 以及该点所在的区域
func Encode(latitude, longitude float64, precision int) (string, *Box) {
    var geohash bytes.Buffer
    var minLat, maxLat float64 = MIN_LATITUDE, MAX_LATITUDE
    var minLng, maxLng float64 = MIN_LONGITUDE, MAX_LONGITUDE
    var mid float64 = 0

    bit, ch, length, isEven := 0, 0, 0, true
    for length < precision {
        if isEven {
            if mid = (minLng + maxLng) / 2; mid < longitude {
                ch |= bits[bit]
                minLng = mid
            } else {
                maxLng = mid
            }
        } else {
            if mid = (minLat + maxLat) / 2; mid < latitude {
                ch |= bits[bit]
                minLat = mid
            } else {
                maxLat = mid
            }
        }

        isEven = !isEven
        if bit < 4 {
            bit++
        } else {
            geohash.WriteByte(base32[ch])
            length, bit, ch = length+1, 0, 0
        }
    }

    b := &Box{
        MinLat: minLat,
        MaxLat: maxLat,
        MinLng: minLng,
        MaxLng: maxLng,
    }

    return geohash.String(), b
}

// 计算该点（latitude, longitude）在精度precision下的邻居 -- 周围8个区域+本身所在区域
// 返回这些区域的geohash值，总共9个
func GetNeighborsByLatLon(latitude, longitude float64, precision int) []string {
    // 本身
    geohash, b := Encode(latitude, longitude, precision)

    return GetNeighborsByGeohash(geohash, precision, b.MinLat, b.MaxLat, b.MinLng, b.MaxLng)
}

func GetNeighborsByGeohash(geohash string, precision int, minLat, maxLat, minLng, maxLng float64 ) []string {

    geohashs := make([]string, 9)

    // 本身
    geohashs[0] = geohash
    b := &Box{
        MinLat: minLat,
        MaxLat: maxLat,
        MinLng: minLng,
        MaxLng: maxLng,
    }

    // 上下左右
    geohashUp, _ := Encode((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2, precision)
    geohashDown, _ := Encode((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2, precision)
    geohashLeft, _ := Encode((b.MinLat+b.MaxLat)/2, (b.MinLng+b.MaxLng)/2-b.Width(), precision)
    geohashRight, _ := Encode((b.MinLat+b.MaxLat)/2, (b.MinLng+b.MaxLng)/2+b.Width(), precision)

    // 四个角
    geohashLeftUp, _ := Encode((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2-b.Width(), precision)
    geohashLeftDown, _ := Encode((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2-b.Width(), precision)
    geohashRightUp, _ := Encode((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2+b.Width(), precision)
    geohashRightDown, _ := Encode((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2+b.Width(), precision)

    geohashs[1], geohashs[2], geohashs[3], geohashs[4] = geohashUp, geohashDown, geohashLeft, geohashRight
    geohashs[5], geohashs[6], geohashs[7], geohashs[8] = geohashLeftUp, geohashLeftDown, geohashRightUp, geohashRightDown

    return geohashs

}

type LatLng struct {
    Lat float64
    Lng float64
}

func DecodeBounds(geohash string) (LatLng, LatLng) {
    var minLat, maxLat float64 = -90, 90
    var minLng, maxLng float64 = -180, 180
    var mid float64 = 0
    isEven := true
    for _, ch := range strings.Split(geohash, "") {
        bit := bytes.Index(base32, []byte(ch))
        i := uint8(4)
        for {
            mask := (bit >> i) & 1;
            if isEven {
                mid = (minLng + maxLng) / 2
                if(mask == 1){
                    minLng = mid
                } else {
                    maxLng = mid
                }
            } else {
                mid = (minLat + maxLat) / 2
                if mask == 1 {
                    minLat = mid
                } else {
                    maxLat = mid
                }
            }
            isEven = !isEven

            if i == 0 {
                break;
            }
            i--
        }
    }
    return LatLng{minLat, minLng}, LatLng{maxLat, maxLng}
}

type Bound struct {
    Min LatLng
    Mid LatLng
    Max LatLng
}

func Decode(geohash string) *Bound {
    latlngMin, latlngMax := DecodeBounds(geohash)
    bound := new(Bound)
    bound.Min = latlngMin
    bound.Max = latlngMax
    bound.Mid = LatLng{
        Lat: (latlngMin.Lat + latlngMax.Lat) / 2,
        Lng: (latlngMin.Lng + latlngMax.Lng) / 2,
    }
    return bound
}

type Neighbors struct {
    Top string
    TopRight string
    Right string
    BottomRight string
    Bottom string
    BottomLeft string
    Left string
    TopLeft string
}

const (
    DIRECTION_TOP        = "top"
    DIRECTION_RIGHT      = "right"
    DIRECTION_BOTTOM     = "bottom"
    DIRECTION_LEFT       = "left"
    BASE32_DIR_TOP       = "p0r21436x8zb9dcf5h7kjnmqesgutwvy"
    BASE32_DIR_RIGHT     = "bc01fg45238967deuvhjyznpkmstqrwx"
    BASE32_DIR_BOTTOM    = "14365h7k9dcfesgujnmqp0r2twvyx8zb"
    BASE32_DIR_LEFT      = "238967debc01fg45kmstqrwxuvhjyznp"
    BASE32_BORDER_RIGHT  = "bcfguvyz"
    BASE32_BORDER_LEFT   = "0145hjnp"
    BASE32_BORDER_TOP    = "prxz"
    BASE32_BORDER_BOTTOM = "028b"
    EVEN                 = "even"
    ODD                  = "odd"
)

var neighbors = map[string]map[string]string {
    "top": {
        "even": BASE32_DIR_TOP,
        "odd":  BASE32_DIR_RIGHT,
    },
    "right": {
        "even": BASE32_DIR_RIGHT,
        "odd":  BASE32_DIR_TOP,
    },
    "bottom": {
        "even": BASE32_DIR_BOTTOM,
        "odd":  BASE32_DIR_LEFT,
    },
    "left": {
        "even": BASE32_DIR_LEFT,
        "odd":  BASE32_DIR_BOTTOM,
    },
}
var borders = map[string]map[string]string {
    "top": {
        "even": BASE32_BORDER_TOP,
        "odd":  BASE32_BORDER_RIGHT,
    },
    "right": {
        "even": BASE32_BORDER_RIGHT,
        "odd":  BASE32_BORDER_TOP,
    },
    "bottom": {
        "even": BASE32_BORDER_BOTTOM,
        "odd":  BASE32_BORDER_LEFT,
    },
    "left": {
        "even": BASE32_BORDER_LEFT,
        "odd":  BASE32_BORDER_BOTTOM,
    },
}

func GetNeighbor(geohash string, direction string) string {
    length := len(geohash)
    last := geohash[(length - 1):]
    oddEven := ODD
    if (length % 2) == 0 {
        oddEven = EVEN
    }
    border := borders[direction][oddEven]
    base := geohash[0:length - 1]
    if strings.Index(border, last) != -1 && 1 < length {
        base = GetNeighbor(base, direction)
    }
    neighbor := neighbors[direction][oddEven]
    return base + string(base32[strings.Index(neighbor, last)])
}

func GetNeighbors(geohash string) Neighbors {
    type result struct { direction string; geohash string }

    worker := func(hash string, direction string, c chan<- result){
        c <- result{direction, GetNeighbor(hash, direction)}
    }

    ch := make(chan result, 8)

    go worker(geohash, DIRECTION_TOP, ch)
    go worker(geohash, DIRECTION_BOTTOM, ch)

    top := <-ch
    bottom := <-ch

    go worker(geohash, DIRECTION_RIGHT, ch)
    go worker(geohash, DIRECTION_LEFT,  ch)
    go worker(top.geohash, DIRECTION_RIGHT, ch)
    go worker(top.geohash, DIRECTION_LEFT,  ch)
    go worker(bottom.geohash, DIRECTION_RIGHT, ch)
    go worker(bottom.geohash, DIRECTION_LEFT, ch)

    right := <-ch
    left := <-ch
    topRight := <-ch
    topLeft := <-ch
    bottomRight := <-ch
    bottomLeft := <-ch

    return Neighbors {
        Top: top.geohash,
        TopRight: topRight.geohash,
        Right: right.geohash,
        BottomRight: bottomRight.geohash,
        Bottom: bottom.geohash,
        BottomLeft: bottomLeft.geohash,
        Left: left.geohash,
        TopLeft: topLeft.geohash,
    }
}

// vim: set ts=8 sw=4 tw=0 :