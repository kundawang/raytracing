package main

import (
	"github.com/fogleman/gg"
	"github.com/liamg/raytracing/raytracer"
)

// Renders a showcase pedestal scene lit by a rectangular area light,
// demonstrating soft penumbra shadows instead of hard point-light edges.
func main() {

	canvasSize := raytracer.Vector{X: 400, Y: 400}

	scene := raytracer.NewScene(raytracer.WithCanvasSize(canvasSize), raytracer.WithBackgroundColour(raytracer.Colour{}))

	// pedestal
	scene.AddObject(raytracer.NewBox(
		raytracer.Vector{X: -1.5, Y: -1.5, Z: 2},
		raytracer.Vector{X: 3, Y: 0.5, Z: 3},
		raytracer.Colour{R: 0.9, G: 0.9, B: 0.9},
		100,
		0.2,
	))
	// exhibit sphere
	scene.AddObject(raytracer.NewSphere(
		raytracer.Vector{X: 0, Y: -0.4, Z: 3.5},
		0.6,
		raytracer.Colour{R: 1, G: 0.2, B: 0.2},
		200,
		0.3,
	))
	// ground plane
	scene.AddObject(raytracer.NewBox(
		raytracer.Vector{X: -20, Y: 0, Z: -10},
		raytracer.Vector{X: 40, Y: 0.1, Z: 40},
		raytracer.Colour{R: 0.8, G: 0.8, B: 0.8},
		10,
		0.1,
	))

	scene.AddLight(raytracer.NewAmbientLight(raytracer.Colour{R: 0.15, G: 0.15, B: 0.15}))
	scene.AddLight(raytracer.NewAreaLight(
		raytracer.Colour{R: 1, G: 1, B: 1},
		raytracer.Vector{X: 1.5, Y: -4, Z: 2},
		raytracer.Vector{X: -0.2, Y: 1, Z: 0.2},
		2, 2,
		64,
	))

	dc := gg.NewContext(int(canvasSize.X), int(canvasSize.Y))

	origin := raytracer.Vector{X: 0, Y: -1.2, Z: -1}

	for x := -int(canvasSize.X) / 2; x < int(canvasSize.X)/2; x++ {
		for y := -int(canvasSize.Y) / 2; y < int(canvasSize.Y)/2; y++ {
			dest := scene.CanvasToViewport(raytracer.Vector{X: float64(x), Y: float64(y)})
			colour, _ := scene.TraceRay(origin, dest, 1, 100, 3)
			dc.SetRGB(colour.R, colour.G, colour.B)
			dc.SetPixel(x+(int(canvasSize.X)/2), (int(canvasSize.Y)/2)+y)
		}
	}

	dc.SavePNG("area.png")
}
