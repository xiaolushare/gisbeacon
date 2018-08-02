/*
 * Copyright (c) 2017 WIFIPIX
 * Author: typdefxiaolu@gmail.com
 * Created Time: 2017/8/14
 *
 */

package wp

import (
    "math"
    "container/list"
    "wifipix.com/gisbeacon/utils"
)

type Polygon struct {
    Path [][2]float64 `json:"polygon"`
}

func NewPolygon(polygon [][2]float64) *Polygon {
    return &Polygon{
        Path: polygon,
    }
}

func (polygon *Polygon)rayCrossesSegment(point [2]float64, a [2]float64, b [2]float64) bool{
    px := point[0]
    py := point[1]
    ax := a[0]
    ay := a[1]
    bx := b[0]
    by := b[1]
    if ay > by {
        ax = b[0]
        ay = b[1]
        bx = a[0]
        by = a[1]
    }
    // alter longitude to cater for 180 degree crossings
    if px < 0 {
        px += 360
    }
    if ax < 0 {
        ax += 360
    }
    if bx < 0 {
        bx += 360
    }
    if py == ay || py == by {
        py += 0.00000001
    }
    if (py > by || py < ay) || (px > math.Max(ax, bx)) {
        return false
    }
    if px < math.Min(ax, bx) {
        return true
    }

    var red , blue float64
    if ax != bx {
        red = (by - ay) / (bx - ax)
    }else {
        red = math.Inf(0)
    }

    if ax != px {
        blue = (py - ay) / (px - ax)
    }else {

        blue = math.Inf(0)
    }

    return blue >= red

}


func (polygon *Polygon)Contains(point [2]float64) bool{
    crossings := 0
    path := polygon.Path
    // for each edge
    for i := 0; i < len(path); i++ {
        a := path[i]
        j := i + 1
        if j >= len(path) {
            j = 0
        }
        b := path[j]
        if polygon.rayCrossesSegment(point, a, b) {
            crossings +=1
        }
    }

    // odd number of crossings?
    return crossings % 2 == 1
}


// 重心 polygon闭合
func (polygon *Polygon) Centroid() [2]float64 {
    var temp float64 = 0
    var area float64 = 0
    var cx float64 = 0
    var cy float64 = 0
    for i :=0; i < len(polygon.Path)-1; i ++{
        temp = polygon.Path[i][0] * polygon.Path[i+1][1] - polygon.Path[i][1]*polygon.Path[i+1][0]
        area += temp
        cx += temp * (polygon.Path[i][0]+polygon.Path[i+1][0])
        cy += temp * (polygon.Path[i][1]+polygon.Path[i+1][1])
    }
    area /= 2
    cx /= (6 * area)
    cy /= (6 * area)
    return [2] float64 {cx, cy}
}


// p inside polygon , 不计算 p > polygon 的情况
func (polygon *Polygon)Inside(p *Polygon) bool {
    flag := true
    for _, point := range p.Path {
        if ! polygon.Contains(point) {
            flag = false
        }
    }
    return flag
}

// 相交但不包含
func (polygon *Polygon)IntersectWithoutInside(p *Polygon) bool{
    inside := false
    outsider := false
    for _, point := range p.Path {
        if ! polygon.Contains(point) {
            outsider = true
        }else {
            inside = true
        }
    }
    return inside && outsider
}

// 相交,包括包含情况
func (polygon *Polygon)Intersect(p *Polygon) bool{
    for _, point := range p.Path {
        if polygon.Contains(point) {
            return true
        }
    }
    return false
}

// TODO polygon geohash set
func GeohashToPolygon(geo string) *Polygon{
    /*
    :param geo: String that represents the geohash.
    :return: Returns a Shapely's Polygon instance that represents the geohash.
    */
    polygon := make([][2]float64, 5)
    minLatLng, maxLatLng := utils.DecodeBounds(geo)
    polygon[0] = [2]float64{minLatLng.Lng, minLatLng.Lat}
    polygon[1] = [2]float64{minLatLng.Lng, maxLatLng.Lat}
    polygon[2] = [2]float64{maxLatLng.Lng, maxLatLng.Lat}
    polygon[3] = [2]float64{maxLatLng.Lng, minLatLng.Lat}
    polygon[4] = [2]float64{minLatLng.Lng, minLatLng.Lat}

    return NewPolygon(polygon)
}

func PolygonToGeohashes(polygon *Polygon, precision int, inner bool) []string {
    /*
    :param polygon: shapely polygon.
    :param precision: int. Geohashes' precision that form resulting polygon.
    :param inner: bool, default 'True'. If false, geohashes that are completely outside from the polygon are ignored.
    :return: set. Set of geohashes that form the polygon.
    */

    inner_geohashes := map[string]bool{}
    outer_geohashes := map[string]bool{}
    testing_geohashes := list.New()
    centroid := polygon.Centroid()
    geohashstr, _ := utils.Encode(centroid[1], centroid[0], precision)
    testing_geohashes.PushFront(geohashstr)

    for testing_geohashes.Len() > 0 {
        e := testing_geohashes.Back()
        current_geohash := e.Value.(string)
        testing_geohashes.Remove(e)

        if _, iok := inner_geohashes[current_geohash]; !iok {
            if _, ook := outer_geohashes[current_geohash]; !ook {
                current_polygon := GeohashToPolygon(current_geohash)
                var condition_inner bool
                var condition_outer bool
                if inner {
                    condition_inner = polygon.Inside(current_polygon)
                } else {
                    condition_outer = polygon.Intersect(current_polygon)
                }
                if condition_inner || condition_outer {
                    if inner {
                        if condition_inner {
                            inner_geohashes[current_geohash] = true
                        } else {
                            outer_geohashes[current_geohash] = true
                        }
                    } else {
                        if condition_outer {
                            inner_geohashes[current_geohash] = true
                        } else {
                            outer_geohashes[current_geohash] = true
                        }
                    }
                    // TODO goroutine
                    //neighbors := utils.GetNeighbors(current_geohash)
                    //for _, neighbor := range []string{
                    //    neighbors.TopLeft, neighbors.Top, neighbors.TopRight,
                    //    neighbors.Left, neighbors.Right,
                    //    neighbors.BottomLeft, neighbors.Bottom, neighbors.BottomRight,
                    //} {
                    //    if _, iok := inner_geohashes[neighbor]; !iok {
                    //        if _, ook := outer_geohashes[neighbor]; !ook {
                    //            testing_geohashes.PushFront(neighbor)
                    //        }
                    //    }
                    //}

                    box := utils.Decode(current_geohash)
                    neighbors := utils.GetNeighborsByGeohash(current_geohash, precision, box.Min.Lat, box.Max.Lat,
                        box.Min.Lng, box.Max.Lng)
                    for _, neighbor := range neighbors {
                        if _, iok := inner_geohashes[neighbor]; !iok {
                            if _, ook := outer_geohashes[neighbor]; !ook {
                                testing_geohashes.PushFront(neighbor)
                            }
                        }
                    }

                }
            }
        }
        //fmt.Println(current_geohash)
    }

    geohashes := []string{}
    for k, _ := range inner_geohashes {
        geohashes = append(geohashes, k)
    }
    return geohashes
}

// vim: set ts=8 sw=4 tw=0 :