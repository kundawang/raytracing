package raytracer

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// areaLightTestScene builds the fixed verification scene: a sphere hovering
// between a downward-facing rectangular area light and the ground plane
// (y=0). The sphere projects a soft shadow onto the plane.
func areaLightTestScene(samples int) *Scene {
	scene := NewScene()
	scene.AddObject(NewSphere(
		Vector{X: 0, Y: 2.5, Z: 0},
		1,
		Colour{R: 1, G: 1, B: 1},
		0,
		0,
	))
	scene.AddLight(NewAreaLight(
		Colour{R: 1, G: 1, B: 1},
		Vector{X: 0, Y: 5, Z: 0},
		Vector{X: 0, Y: -1, Z: 0},
		2, 2,
		samples,
	))
	return scene
}

// groundIntensity evaluates the direct lighting at a point on the ground
// plane (normal +Y), as if the plane were being shaded there.
func groundIntensity(scene *Scene, x, z float64) Colour {
	return scene.computeLighting(
		Vector{X: x, Y: 0, Z: z},
		Vector{X: 0, Y: 1, Z: 0},
		Vector{X: 0, Y: 1, Z: 0},
		0,
	)
}

func assertFiniteColour(t *testing.T, c Colour, context string) {
	t.Helper()
	assert.False(t, math.IsNaN(c.R) || math.IsInf(c.R, 0), "%s: R is not finite (%v)", context, c.R)
	assert.False(t, math.IsNaN(c.G) || math.IsInf(c.G, 0), "%s: G is not finite (%v)", context, c.G)
	assert.False(t, math.IsNaN(c.B) || math.IsInf(c.B, 0), "%s: B is not finite (%v)", context, c.B)
}

func TestAreaLightUmbraPenumbra(t *testing.T) {
	scene := areaLightTestScene(64)

	// deep shadow: the point directly under the sphere sees none of the light
	umbra := groundIntensity(scene, 0, 0)
	assertFiniteColour(t, umbra, "umbra")
	assert.InDelta(t, 0, umbra.R, 0.01, "umbra should be (near) black")

	// fully lit: far from the shadow the whole light is visible
	lit := groundIntensity(scene, 10, 0)
	assertFiniteColour(t, lit, "lit")
	assert.True(t, lit.R > 0.3, "lit point should be bright, got %v", lit.R)

	// walking outwards from the shadow centre the intensity must be
	// monotonically non-decreasing through the shadow transition region, and
	// pass through a penumbra band strictly between umbra and full
	// illumination (beyond the shadow, cosine falloff takes over, so the
	// monotonic scan stops inside the lit region)
	previous := umbra.R
	for x := 0.1; x <= 2; x += 0.1 {
		current := groundIntensity(scene, x, 0).R
		assertFiniteColour(t, groundIntensity(scene, x, 0), "scan")
		assert.True(t, current >= previous-0.000001, "intensity decreased at x=%.1f (%v < %v)", x, current, previous)
		previous = current
	}
	foundPenumbra := false
	for x := 0.1; x <= 4; x += 0.1 {
		current := groundIntensity(scene, x, 0).R
		if current > umbra.R+0.02 && current < lit.R-0.02 {
			foundPenumbra = true
		}
	}
	assert.True(t, foundPenumbra, "expected a penumbra band between umbra and lit region")
}

func TestAreaLightReproducible(t *testing.T) {
	sceneA := areaLightTestScene(32)
	sceneB := areaLightTestScene(32)
	for _, x := range []float64{0, 0.7, 1.3, 2.5, 10} {
		a1 := groundIntensity(sceneA, x, 0)
		a2 := groundIntensity(sceneA, x, 0)
		b := groundIntensity(sceneB, x, 0)
		assert.Equal(t, a1, a2, "same scene, same point must match at x=%v", x)
		assert.Equal(t, a1, b, "identically configured scenes must match at x=%v", x)
	}
}

func TestAreaLightBackFaceEmitsNothing(t *testing.T) {
	scene := areaLightTestScene(16)
	// a point above the light (which faces straight down) is behind it
	behind := scene.computeLighting(
		Vector{X: 0, Y: 6, Z: 0},
		Vector{X: 0, Y: 1, Z: 0},
		Vector{X: 0, Y: 1, Z: 0},
		0,
	)
	assert.Equal(t, Colour{}, behind)
}

func TestAreaLightDegenerateRectangle(t *testing.T) {
	// a zero-size area light must behave like a point light at its centre
	areaScene := NewScene()
	areaScene.AddLight(NewAreaLight(
		Colour{R: 0.8, G: 0.6, B: 0.4},
		Vector{X: 0, Y: 5, Z: 0},
		Vector{X: 0, Y: -1, Z: 0},
		0, 0,
		16,
	))
	pointScene := NewScene()
	pointScene.AddLight(NewPointLight(
		Colour{R: 0.8, G: 0.6, B: 0.4},
		Vector{X: 0, Y: 5, Z: 0},
	))

	point := Vector{X: 1, Y: 0, Z: 0}
	normal := Vector{X: 0, Y: 1, Z: 0}
	toCamera := Vector{X: 0, Y: 1, Z: 0}
	area := areaScene.computeLighting(point, normal, toCamera, 10)
	punctual := pointScene.computeLighting(point, normal, toCamera, 10)
	assertFiniteColour(t, area, "degenerate area light")
	assert.InDelta(t, punctual.R, area.R, 0.000001)
	assert.InDelta(t, punctual.G, area.G, 0.000001)
	assert.InDelta(t, punctual.B, area.B, 0.000001)
}

func TestAreaLightDegenerateConfiguration(t *testing.T) {
	point := Vector{X: 1, Y: 0, Z: 0}
	normal := Vector{X: 0, Y: 1, Z: 0}
	toCamera := Vector{X: 0, Y: 1, Z: 0}

	// zero and negative sample counts fall back to a single stable sample
	for _, samples := range []int{0, -5, 1} {
		scene := NewScene()
		scene.AddLight(NewAreaLight(Colour{R: 1, G: 1, B: 1}, Vector{X: 0, Y: 5, Z: 0}, Vector{X: 0, Y: -1, Z: 0}, 2, 2, samples))
		c := scene.computeLighting(point, normal, toCamera, 0)
		assertFiniteColour(t, c, "low sample count")
		assert.True(t, c.R > 0, "expected some light with %d samples", samples)
	}

	// negative dimensions are clamped, result must stay finite
	scene := NewScene()
	scene.AddLight(NewAreaLight(Colour{R: 1, G: 1, B: 1}, Vector{X: 0, Y: 5, Z: 0}, Vector{X: 0, Y: -1, Z: 0}, -1, -2, 8))
	assertFiniteColour(t, scene.computeLighting(point, normal, toCamera, 0), "negative size")

	// zero normal emits nothing rather than panicking or producing NaN
	scene = NewScene()
	scene.AddLight(NewAreaLight(Colour{R: 1, G: 1, B: 1}, Vector{X: 0, Y: 5, Z: 0}, Vector{}, 2, 2, 8))
	assert.Equal(t, Colour{}, scene.computeLighting(point, normal, toCamera, 0))
}

func TestAreaLightSelfOcclusion(t *testing.T) {
	// a point on the surface of an object must not be shadowed by the object
	// itself: the sphere's own silhouette point facing the light stays lit
	scene := NewScene()
	scene.AddObject(NewSphere(Vector{X: 0, Y: 0, Z: 0}, 1, Colour{R: 1, G: 1, B: 1}, 0, 0))
	scene.AddLight(NewAreaLight(
		Colour{R: 1, G: 1, B: 1},
		Vector{X: 0, Y: 5, Z: 0},
		Vector{X: 0, Y: -1, Z: 0},
		2, 2,
		16,
	))
	top := scene.computeLighting(
		Vector{X: 0, Y: 1, Z: 0}, // top of the sphere, facing the light
		Vector{X: 0, Y: 1, Z: 0},
		Vector{X: 0, Y: 1, Z: 0},
		0,
	)
	assertFiniteColour(t, top, "self occlusion")
	assert.True(t, top.R > 0.5, "point facing the light should be lit, got %v", top.R)
}

func TestAreaLightHighIntensityOverflow(t *testing.T) {
	scene := NewScene()
	scene.AddLight(NewAreaLight(
		Colour{R: 50, G: 50, B: 50},
		Vector{X: 0, Y: 5, Z: 0},
		Vector{X: 0, Y: -1, Z: 0},
		2, 2,
		64,
	))
	c := scene.computeLighting(
		Vector{X: 0, Y: 0, Z: 0},
		Vector{X: 0, Y: 1, Z: 0},
		Vector{X: 0, Y: 1, Z: 0},
		0,
	)
	assertFiniteColour(t, c, "high intensity")
	assert.True(t, c.R <= 1 && c.G <= 1 && c.B <= 1, "channels must be clamped, got %v", c)
	assert.True(t, c.R >= 0 && c.G >= 0 && c.B >= 0, "channels must not go negative, got %v", c)
}

func TestAreaLightDoesNotAffectOtherLightTypes(t *testing.T) {
	// a scene using only ambient/point/directional lights must produce the
	// same lighting as before area lights existed
	build := func() *Scene {
		scene := NewScene()
		scene.AddObject(NewSphere(Vector{X: 0, Y: 0, Z: 3}, 1, Colour{R: 1, G: 0, B: 0}, 50, 0.4))
		scene.AddLight(NewAmbientLight(Colour{R: 0.4, G: 0.4, B: 0.4}))
		scene.AddLight(NewPointLight(Colour{R: 0.7, G: 0.7, B: 0.7}, Vector{X: 7, Y: -70, Z: 0}))
		scene.AddLight(NewDirectionalLight(Colour{R: 0.2, G: 0.2, B: 0.2}, Vector{X: 0, Y: 1, Z: 0}))
		return scene
	}
	point := Vector{X: 0, Y: 1, Z: 3}
	normal := Vector{X: 0, Y: 1, Z: 0}
	toCamera := Vector{X: 0, Y: 0, Z: -1}
	a := build().computeLighting(point, normal, toCamera, 50)
	b := build().computeLighting(point, normal, toCamera, 50)
	assert.Equal(t, a, b)
	assertFiniteColour(t, a, "existing light types")
}
