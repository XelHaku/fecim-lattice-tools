// Package render3d provides a pure-Go software 3D renderer for multi-layer
// FeCIM array stack visualization. It renders stacked crossbar layers with
// isometric projection into standard Go images suitable for Fyne canvas.Raster.
//
// No external 3D libraries are used. All math is standard library only.
package render3d

import "math"

// Vec3 represents a 3D vector or point.
type Vec3 struct{ X, Y, Z float64 }

// Add returns the component-wise sum of two vectors.
func (v Vec3) Add(w Vec3) Vec3 {
	return Vec3{v.X + w.X, v.Y + w.Y, v.Z + w.Z}
}

// Sub returns the component-wise difference of two vectors.
func (v Vec3) Sub(w Vec3) Vec3 {
	return Vec3{v.X - w.X, v.Y - w.Y, v.Z - w.Z}
}

// Scale returns the vector scaled by a scalar.
func (v Vec3) Scale(s float64) Vec3 {
	return Vec3{v.X * s, v.Y * s, v.Z * s}
}

// Dot returns the dot product of two vectors.
func (v Vec3) Dot(w Vec3) float64 {
	return v.X*w.X + v.Y*w.Y + v.Z*w.Z
}

// Length returns the Euclidean length of the vector.
func (v Vec3) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Normalize returns a unit-length vector in the same direction.
// Returns a zero vector if the input has zero length.
func (v Vec3) Normalize() Vec3 {
	l := v.Length()
	if l < 1e-15 {
		return Vec3{}
	}
	return Vec3{v.X / l, v.Y / l, v.Z / l}
}

// Mat4 is a 4x4 matrix stored in column-major order.
// Index layout: [col*4 + row], matching OpenGL convention.
//
//	M[0] M[4] M[8]  M[12]
//	M[1] M[5] M[9]  M[13]
//	M[2] M[6] M[10] M[14]
//	M[3] M[7] M[11] M[15]
type Mat4 [16]float64

// Identity returns the 4x4 identity matrix.
func Identity() Mat4 {
	var m Mat4
	m[0] = 1
	m[5] = 1
	m[10] = 1
	m[15] = 1
	return m
}

// Multiply returns the product of two 4x4 matrices (a * b).
func (a Mat4) Multiply(b Mat4) Mat4 {
	var m Mat4
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			sum := 0.0
			for k := 0; k < 4; k++ {
				sum += a[k*4+row] * b[col*4+k]
			}
			m[col*4+row] = sum
		}
	}
	return m
}

// Project transforms a 3D point to 2D screen coordinates using the matrix.
// The result is the (x, y) position after projection (w-division ignored for
// orthographic/isometric projection where w=1).
func (m Mat4) Project(v Vec3) (screenX, screenY float64) {
	x := m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12]
	y := m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13]
	return x, y
}

// TransformVec3 applies the full 4x4 transform to a 3D point (with w=1)
// and returns the resulting 3D point. Useful for depth sorting.
func (m Mat4) TransformVec3(v Vec3) Vec3 {
	x := m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12]
	y := m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13]
	z := m[2]*v.X + m[6]*v.Y + m[10]*v.Z + m[14]
	return Vec3{x, y, z}
}

// RotateY returns a rotation matrix around the Y axis by the given angle in radians.
func RotateY(angle float64) Mat4 {
	c := math.Cos(angle)
	s := math.Sin(angle)
	m := Identity()
	m[0] = c
	m[8] = s
	m[2] = -s
	m[10] = c
	return m
}

// RotateX returns a rotation matrix around the X axis by the given angle in radians.
func RotateX(angle float64) Mat4 {
	c := math.Cos(angle)
	s := math.Sin(angle)
	m := Identity()
	m[5] = c
	m[9] = -s
	m[6] = s
	m[10] = c
	return m
}

// ScaleMat returns a uniform scale matrix.
func ScaleMat(sx, sy, sz float64) Mat4 {
	m := Identity()
	m[0] = sx
	m[5] = sy
	m[10] = sz
	return m
}

// Translate returns a translation matrix.
func Translate(tx, ty, tz float64) Mat4 {
	m := Identity()
	m[12] = tx
	m[13] = ty
	m[14] = tz
	return m
}

// NewIsometricProjection creates an isometric camera matrix.
//
// Parameters:
//   - azimuth: rotation around Y-axis (radians). 0 looks along +Z.
//   - elevation: tilt angle from horizontal (radians). pi/6 is standard isometric.
//   - scale: uniform scale factor mapping world units to pixels.
func NewIsometricProjection(azimuth, elevation, scale float64) Mat4 {
	// Build the transform: Scale * RotateX(elevation) * RotateY(azimuth)
	ry := RotateY(azimuth)
	rx := RotateX(elevation)
	s := ScaleMat(scale, -scale, scale) // Negate Y to flip from math coords to screen coords (Y down)
	return s.Multiply(rx).Multiply(ry)
}
