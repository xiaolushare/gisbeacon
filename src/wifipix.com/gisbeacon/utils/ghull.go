/*
 * Copyright (c) 2018 typedef.pw
 * Author: typedefxiaolu@gamil.com
 * Created Time: 2018/7/27
 *
 */

package utils

import "sort"

type Point struct {
    X, Y float64
}

type Points []Point

func (points Points) Swap(i, j int) {
    points[i], points[j] = points[j], points[i]
}

func (points Points) Less(i, j int) bool {
    if points[i].X == points[j].X {
        return points[i].Y < points[j].Y
    }
    return points[i].X < points[j].X
}

func (points Points) Len() int {
    return len(points)
}

func CrossProduct(a, b, c Point) float64 {
    return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

type Hull struct {
    points Points
}

func lowerHull(points Points, size int) chan Points {
    result := make(chan Points)
    go func() {
        var lower Points
        count := 0
        for i := 0; i < size; i++ {
            for count >= 2 && CrossProduct(lower[count-2], lower[count-1], points[i]) <= 0 {
                count--
                lower = lower[:count]
            }
            count++
            lower = append(lower, points[i])
        }
        result <- lower[:len(lower)-1]
    }()
    return result
}

func upperHull(points Points, size int) chan Points {
    result := make(chan Points)
    go func() {
        var upper Points
        count := 0
        for i := size - 1; i >= 0; i-- {
            for count >= 2 && CrossProduct(upper[count-2], upper[count-1], points[i]) <= 0 {
                count--
                upper = upper[:count]
            }
            count++
            upper = append(upper, points[i])
        }
        result <- upper[:len(upper)-1]
    }()
    return result
}

func convexHull(points Points) *Hull {
    var result Points
    size := len(points)
    if size < 3 {
        return &Hull{result}
    }

    sort.Sort(points)

    lower := <-lowerHull(points, size)
    upper := <-upperHull(points, size)

    result = append(result, lower...)
    result = append(result, upper...)
    return &Hull{result}
}

func GetConvexHullPolygon(points [][]float64)(result [][]float64) {
    ps := make(Points, 0)
    for _, point := range points {
        ps = append(ps, Point{X: point[0], Y: point[1]})
    }
    hull := convexHull(ps)
    //result := make([][]float64, 0)
    for _, point := range hull.points {
        result = append(result, []float64{point.X, point.Y})
    }
    return result
}

// vim: set ts=8 sw=4 tw=0 :