package geo

import (
	"math"

	"github.com/paulmach/orb"
)

func PointInPolygon(point orb.Point, polygon orb.Polygon) bool {
	eps := 1e-14
	dist := func(p, a, b orb.Point) float64 {
		ax, ay := a[0], a[1]
		bx, by := b[0], b[1]
		px, py := p[0], p[1]
		dx, dy := bx-ax, by-ay
		if dx == 0 && dy == 0 {
			return math.Hypot(px-ax, py-ay)
		}
		t := ((px-ax)*dx + (py-ay)*dy) / (dx*dx + dy*dy)
		if t < 0 {
			return math.Hypot(px-ax, py-ay)
		} else if t > 1 {
			return math.Hypot(px-bx, py-by)
		}
		cx := ax + t*dx
		cy := ay + t*dy
		return math.Hypot(px-cx, py-cy)
	}
	orientation := func(r orb.Ring) float64 {
		var sum float64
		for i := 0; i < len(r)-1; i++ {
			sum += (r[i+1][0] - r[i][0]) * (r[i+1][1] + r[i][1])
		}
		return sum
	}
	winding := func(pt orb.Point, r orb.Ring) float64 {
		var total float64
		for i := 0; i < len(r)-1; i++ {
			x1, y1 := r[i][0]-pt[0], r[i][1]-pt[1]
			x2, y2 := r[i+1][0]-pt[0], r[i+1][1]-pt[1]
			if dist(pt, r[i], r[i+1]) <= eps {
				return 2 * math.Pi
			}
			dot := x1*x2 + y1*y2
			crs := x1*y2 - y1*x2
			total += math.Atan2(crs, dot)
		}
		return total
	}
	var sum float64
	for _, ring := range polygon {
		if len(ring) < 4 || ring[0] != ring[len(ring)-1] {
			continue
		}
		o := orientation(ring)
		w := winding(point, ring)
		if o > 0 {
			sum += w
		} else {
			sum -= w
		}
	}
	return math.Abs(sum) > math.Pi
}
