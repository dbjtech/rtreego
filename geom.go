// Copyright 2012 Daniel Connelly.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// 为了简单表述，下面注释中把所有包络框或者叫边界框，叫做最小外接矩形边框

package rtreego

import (
	"fmt"
	"math"
	"strings"
)

// DimError represents a failure due to mismatched dimensions.
//
// 表示由于维度不匹配导致的失败。
type DimError struct {
	Expected int
	Actual   int
}

func (err DimError) Error() string {
	return "rtreego: dimension mismatch"
}

// DistError is an improper distance measurement.  It implements the error
// and is generated when a distance-related assertion fails.
//
// 表示不正确的距离测量。它实现了 `error` 接口，并在距离相关的断言失败时生成。
type DistError float32

func (err DistError) Error() string {
	return "rtreego: improper distance"
}

// Point represents a point in n-dimensional Euclidean space.
//
// 表示 n 维欧几里得空间中的一个点。
type Point []float32

func (p Point) Copy() Point {
	result := make(Point, len(p))
	copy(result, p)
	return result
}

// Dist computes the Euclidean distance between two points p and q.
//
// 计算 p 和 q 两点之间的欧几里得距离。就是两点间 每个维度上的差值平方和的平方根
func (p Point) dist(q Point) float32 {
	if len(p) != len(q) {
		panic(DimError{len(p), len(q)})
	}
	sum := float32(0.0)
	for i := range p {
		dx := p[i] - q[i]
		sum += dx * dx
	}
	return float32(math.Sqrt(float64(sum)))
}

// minDist computes the square of the distance from a point to a rectangle.
// If the point is contained in the rectangle then the distance is zero.
//
// Implemented per Definition 2 of "Nearest Neighbor Queries" by
// N. Roussopoulos, S. Kelley and F. Vincent, ACM SIGMOD, pages 71-79, 1995.
//
// 计算从一个点到一个矩形的距离的平方。如果该点包含在矩形内，则距离为零。
func (p Point) minDist(r Rect) float64 {
	if len(p) != len(r.p) {
		panic(DimError{len(p), len(r.p)})
	}

	sum := 0.0
	for i, pi := range p {
		if pi < r.p[i] {
			d := pi - r.p[i]
			sum += float64(d) * float64(d)
		} else if pi > r.q[i] {
			d := pi - r.q[i]
			sum += float64(d) * float64(d)
		} else {
			sum += 0
		}
	}
	return sum
}

// minMaxDist computes the minimum of the maximum distances from p to points
// on r.  If r is the bounding box of some geometric objects, then there is
// at least one object contained in r within minMaxDist(p, r) of p.
//
// Implemented per Definition 4 of "Nearest Neighbor Queries" by
// N. Roussopoulos, S. Kelley and F. Vincent, ACM SIGMOD, pages 71-79, 1995.
//
// 计算从`p`到`r`上各点的最大距离中的最小值。如果`r`是某些几何对象的边界框，那么在`r`中至少有一个对象位于`p`的minMaxDist(p, r)范围内
func (p Point) minMaxDist(r Rect) float64 {
	if len(p) != len(r.p) {
		panic(DimError{len(p), len(r.p)})
	}

	// by definition, MinMaxDist(p, r) =
	// min{1<=k<=n}(|pk - rmk|^2 + sum{1<=i<=n, i != k}(|pi - rMi|^2))
	// where rmk and rMk are defined as follows:

	rm := func(k int) float32 {
		if p[k] <= (r.p[k]+r.q[k])/2 {
			return r.p[k]
		}
		return r.q[k]
	}

	rM := func(k int) float32 {
		if p[k] >= (r.p[k]+r.q[k])/2 {
			return r.p[k]
		}
		return r.q[k]
	}

	// This formula can be computed in linear time by precomputing
	// S = sum{1<=i<=n}(|pi - rMi|^2).

	S := 0.0
	for i := range p {
		d := p[i] - rM(i)
		S += float64(d) * float64(d)
	}

	// Compute MinMaxDist using the precomputed S.
	min := math.MaxFloat64
	for k := range p {
		d1 := p[k] - rM(k)
		d2 := p[k] - rm(k)
		d := S - float64(d1)*float64(d1) + float64(d2)*float64(d2)
		if d < min {
			min = d
		}
	}

	return min
}

// Rect represents a subset of n-dimensional Euclidean space of the form
// [a1, b1] x [a2, b2] x ... x [an, bn], where ai < bi for all 1 <= i <= n.
type Rect struct {
	p, q     Point // Enforced by NewRect: p[i] <= q[i] for all i.
	traceBox []Rect
}

// PointCoord returns the coordinate of the point of the rectangle at i
//
// 返回矩形在索引 i 处的坐标。
func (r Rect) PointCoord(i int) float32 {
	return r.p[i]
}

// LengthsCoord returns the coordinate of the lengths of the rectangle at i
func (r Rect) LengthsCoord(i int) float32 {
	return r.q[i] - r.p[i]
}

// Equal returns true if the two rectangles are equal
func (r Rect) Equal(other Rect) bool {
	for i, e := range r.p {
		if e != other.p[i] {
			return false
		}
	}
	for i, e := range r.q {
		if e != other.q[i] {
			return false
		}
	}
	return true
}

func (r Rect) String() string {
	s := make([]string, len(r.p))
	for i, a := range r.p {
		b := r.q[i]
		s[i] = fmt.Sprintf("[%.2f, %.2f]", a, b)
	}
	return strings.Join(s, "x")
}

func (r Rect) AppendTraceBox(bb Rect) Spatial {
	r.traceBox = append(r.traceBox, bb)
	return r
}

func (r Rect) Bounds() Rect {
	return r
}

func (r Rect) StartEnd() []Point {
	return []Point{r.p, r.q}
}

// NewRect constructs and returns a pointer to a Rect given a corner point and
// the lengths of each dimension.  The point p should be the most-negative point
// on the rectangle (in every dimension) and every length should be positive.
func NewRect(p Point, lengths []float32) (r Rect, err error) {
	r.p = p
	if len(p) != len(lengths) {
		err = &DimError{len(p), len(lengths)}
		return
	}
	r.q = make([]float32, len(p))
	for i := range p {
		if lengths[i] <= 0 {
			err = DistError(lengths[i])
			return
		}
		r.q[i] = p[i] + lengths[i]
	}
	return
}

// NewRectFromPoints constructs and returns a pointer to a Rect given a corner points.
func NewRectFromPoints(minPoint, maxPoint Point) (r Rect, err error) {
	if len(minPoint) != len(maxPoint) {
		err = &DimError{len(minPoint), len(maxPoint)}
		return
	}

	// check that min and max point coordinates require swapping
	copied := false
	for i, p := range minPoint {
		if minPoint[i] > maxPoint[i] {
			if !copied {
				minPoint = minPoint.Copy()
				maxPoint = maxPoint.Copy()
				copied = true
			}
			minPoint[i] = maxPoint[i]
			maxPoint[i] = p
		}
	}

	r = Rect{p: minPoint, q: maxPoint}
	return
}

// Size computes the measure of a rectangle (the product of its side lengths).
//
// 返回矩形的测度如面积/体积等。
func (r Rect) Size() float32 {
	size := float32(1.0)
	for i, a := range r.p {
		b := r.q[i]
		size *= b - a
	}
	return size
}

// margin computes the sum of the edge lengths of a rectangle.
func (r Rect) margin() float32 {
	// The number of edges in an n-dimensional rectangle is n * 2^(n-1)
	// (http://en.wikipedia.org/wiki/Hypercube_graph).  Thus the number
	// of edges of length (ai - bi), where the rectangle is determined
	// by p = (a1, a2, ..., an) and q = (b1, b2, ..., bn), is 2^(n-1).
	//
	// The margin of the rectangle, then, is given by the formula
	// 2^(n-1) * [(b1 - a1) + (b2 - a2) + ... + (bn - an)].
	dim := len(r.p)
	sum := float32(0.0)
	for i, a := range r.p {
		b := r.q[i]
		sum += b - a
	}
	return float32(math.Pow(2, float64(dim-1)) * float64(sum))
}

// containsPoint tests whether p is located inside or on the boundary of r.
func (r Rect) containsPoint(p Point) bool {
	if len(p) != len(r.p) {
		panic(DimError{len(r.p), len(p)})
	}

	for i, a := range p {
		// p is contained in (or on) r if and only if p <= a <= q for
		// every dimension.
		if a < r.p[i] || a > r.q[i] {
			return false
		}
	}

	return true
}

// containsRect tests whether r2 is is located inside r1.
func (r Rect) containsRect(r2 Rect) bool {
	if len(r.p) != len(r2.p) {
		panic(DimError{len(r.p), len(r2.p)})
	}

	for i, a1 := range r.p {
		b1, a2, b2 := r.q[i], r2.p[i], r2.q[i]
		// enforced by constructor: a1 <= b1 and a2 <= b2.
		// so containment holds if and only if a1 <= a2 <= b2 <= b1
		// for every dimension.
		// float32精度不够高，精确到第六位就行了
		if a1-a2 > 0.000001 || b2-b1 > 0.000001 {
			return false
		}
	}

	return true
}

// intersect computes the intersection of two rectangles.  If no intersection
// exists, the intersection is nil.
func intersect(r1, r2 Rect) bool {
	dim := len(r1.p)
	if len(r2.p) != dim {
		panic(DimError{dim, len(r2.p)})
	}

	// There are four cases of overlap:
	//
	//     1.  a1------------b1
	//              a2------------b2
	//              p--------q
	//
	//     2.       a1------------b1
	//         a2------------b2
	//              p--------q
	//
	//     3.  a1-----------------b1
	//              a2-------b2
	//              p--------q
	//
	//     4.       a1-------b1
	//         a2-----------------b2
	//              p--------q
	//
	// Thus there are only two cases of non-overlap:
	//
	//     1. a1------b1
	//                    a2------b2
	//
	//     2.             a1------b1
	//        a2------b2
	//
	// Enforced by constructor: a1 <= b1 and a2 <= b2.  So we can just
	// check the endpoints.

	for i := range r1.p {
		a1, b1, a2, b2 := r1.p[i], r1.q[i], r2.p[i], r2.q[i]
		if b2 <= a1 || b1 <= a2 {
			return false
		}
	}
	return true
}

// ToRect constructs a rectangle containing p with side lengths 2*tol.
func (p Point) ToRect(tol float32) Rect {
	dim := len(p)
	a, b := make([]float32, dim), make([]float32, dim)
	for i := range p {
		a[i] = p[i] - tol
		b[i] = p[i] + tol
	}
	return Rect{a, b, nil}
}

// boundingBox constructs the smallest rectangle containing both r1 and r2.
//
// 构建出同时包含 r1 和 r2 的最小矩形。
func boundingBox(r1, r2 Rect) (bb Rect) {
	dim := len(r1.p)
	bb.p = make([]float32, dim)
	bb.q = make([]float32, dim)
	if len(r2.p) != dim {
		panic(DimError{dim, len(r2.p)})
	}
	for i := 0; i < dim; i++ {
		if r1.p[i] <= r2.p[i] {
			bb.p[i] = r1.p[i]
		} else {
			bb.p[i] = r2.p[i]
		}
		if r1.q[i] <= r2.q[i] {
			bb.q[i] = r2.q[i]
		} else {
			bb.q[i] = r1.q[i]
		}
	}
	return
}
