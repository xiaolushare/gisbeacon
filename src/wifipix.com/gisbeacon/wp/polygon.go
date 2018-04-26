/*
 * Copyright (c) 2017 WIFIPIX
 * Author: typdefxiaolu@gmail.com
 * Created Time: 2017/8/14
 *
 */

package wp

import (
    "math"
)

type Polygon struct {
    Path [][]float64 `json:"polygon"`
}

func NewPolygon(polygon [][]float64) *Polygon {
    return &Polygon{
        Path: polygon,
    }
}

func (polygon *Polygon)rayCrossesSegment(point [2]float64, a []float64, b []float64) bool{
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

// vim: set ts=8 sw=4 tw=0 :