package raytracer

import (
	"math"
	"math/rand"
)

type Light struct {
	direction Vector
	position  Vector
	intensity Colour
	lightType LightType

	// area lights only
	normal  Vector   // unit emission direction of the rectangle
	u       Vector   // unit basis vector along the rectangle width
	v       Vector   // unit basis vector along the rectangle height
	samples []Vector // precomputed jittered offsets from the rectangle centre
}

type LightType uint8

const (
	Ambient LightType = iota
	Point
	Directional
	Area
)

// areaLightSeed is a fixed seed for area light sample jitter so that renders
// are reproducible for a given light configuration.
const areaLightSeed = 1337

func NewAmbientLight(intensity Colour) Light {
	return Light{
		lightType: Ambient,
		intensity: intensity,
	}
}

func NewPointLight(intensity Colour, point Vector) Light {
	return Light{
		lightType: Point,
		intensity: intensity,
		position:  point,
	}
}

func NewDirectionalLight(intensity Colour, direction Vector) Light {
	return Light{
		lightType: Directional,
		intensity: intensity,
		direction: direction,
	}
}

// NewAreaLight creates a rectangular area light centred at position, emitting
// from the face that normal points away from (the back face emits nothing).
// width and height are the rectangle dimensions; negative values are treated
// as zero, and a zero-size rectangle behaves like a point light. samples is
// the requested number of shadow samples - the actual count is rounded up to
// the next square number so samples can be stratified on a jittered grid.
// Sample positions are generated from a fixed seed, so results are
// reproducible across runs. A zero normal produces a light that emits nothing.
func NewAreaLight(intensity Colour, position Vector, normal Vector, width float64, height float64, samples int) Light {
	light := Light{
		lightType: Area,
		intensity: intensity,
		position:  position,
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	if normal.Length() == 0 {
		// degenerate orientation: leave samples nil so the light emits nothing
		return light
	}
	light.normal = normal.Normalise()

	// build an orthonormal basis (u, v) spanning the rectangle plane
	arbitrary := Vector{X: 0, Y: 1, Z: 0}
	if math.Abs(light.normal.DotProduct(arbitrary)) > 0.99 {
		arbitrary = Vector{X: 1, Y: 0, Z: 0}
	}
	light.u = light.normal.CrossProduct(arbitrary).Normalise()
	light.v = light.normal.CrossProduct(light.u).Normalise()

	if samples < 1 {
		samples = 1
	}
	grid := int(math.Ceil(math.Sqrt(float64(samples))))
	rng := rand.New(rand.NewSource(areaLightSeed))
	light.samples = make([]Vector, 0, grid*grid)
	for x := 0; x < grid; x++ {
		for y := 0; y < grid; y++ {
			// stratified jittered sample in [0,1)^2, mapped onto the rectangle
			jx := (float64(x) + rng.Float64()) / float64(grid)
			jy := (float64(y) + rng.Float64()) / float64(grid)
			offset := light.u.MultiplyN((jx - 0.5) * width).Add(light.v.MultiplyN((jy - 0.5) * height))
			light.samples = append(light.samples, offset)
		}
	}
	return light
}
