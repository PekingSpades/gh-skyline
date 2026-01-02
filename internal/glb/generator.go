// Package glb provides functionality for generating GLB (binary glTF) files
// with vertex colors from GitHub contribution data.
package glb

import (
	"fmt"

	"github.com/github/gh-skyline/internal/errors"
	"github.com/github/gh-skyline/internal/logger"
	"github.com/github/gh-skyline/internal/stl/geometry"
	"github.com/github/gh-skyline/internal/types"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

// GenerateGLB creates a colored 3D model from GitHub contribution data and writes it to a GLB file.
func GenerateGLB(contributions [][]types.ContributionDay, outputPath, username string, year int) error {
	contributionsRange := [][][]types.ContributionDay{contributions}
	return GenerateGLBRange(contributionsRange, outputPath, username, year, year)
}

// GenerateGLBRange creates a colored 3D model from multiple years of GitHub contribution data.
func GenerateGLBRange(contributions [][][]types.ContributionDay, outputPath, username string, startYear, endYear int) error {
	log := logger.GetLogger()
	if err := log.Debug("Starting GLB generation for user %s, years %d-%d", username, startYear, endYear); err != nil {
		return errors.Wrap(err, "failed to log debug message")
	}

	if len(contributions) == 0 || len(contributions[0]) == 0 {
		return errors.New(errors.ValidationError, "contributions data cannot be empty", nil)
	}

	// Find global max contribution across all years
	maxContribution := findMaxContributionsAcrossYears(contributions)

	// Calculate dimensions
	yearCount := len(contributions)
	baseWidth, baseDepth := geometry.CalculateMultiYearDimensions(yearCount)

	// Format year string
	yearStr := fmt.Sprintf("%d", endYear)
	if startYear != endYear {
		yearStr = fmt.Sprintf("%04d-%02d", startYear, endYear%100)
	}

	// Create glTF document
	doc := gltf.NewDocument()

	// Generate all geometry with colors
	positions, colors, indices := generateAllGeometry(contributions, maxContribution, baseWidth, baseDepth, username, yearStr)

	if len(positions) == 0 {
		return errors.New(errors.ValidationError, "no geometry generated", nil)
	}

	// Write geometry data to document using modeler
	positionAccessor := modeler.WritePosition(doc, positions)
	colorAccessor := modeler.WriteColor(doc, colors)
	indicesAccessor := modeler.WriteIndices(doc, indices)

	// Create mesh with primitive
	doc.Meshes = []*gltf.Mesh{{
		Name: fmt.Sprintf("%s-skyline-%d", username, endYear),
		Primitives: []*gltf.Primitive{{
			Indices: gltf.Index(indicesAccessor),
			Attributes: gltf.PrimitiveAttributes{
				gltf.POSITION: positionAccessor,
				gltf.COLOR_0:  colorAccessor,
			},
			Mode: gltf.PrimitiveTriangles,
		}},
	}}

	// Create node referencing the mesh
	doc.Nodes = []*gltf.Node{{
		Name: "Skyline",
		Mesh: gltf.Index(0),
	}}

	// Add node to scene
	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, 0)

	// Save as binary GLB
	if err := gltf.SaveBinary(doc, outputPath); err != nil {
		return errors.New(errors.IOError, "failed to write GLB file", err)
	}

	if err := log.Info("GLB file written successfully to: %s", outputPath); err != nil {
		return errors.Wrap(err, "failed to log info message")
	}

	return nil
}

// generateAllGeometry generates positions, colors, and indices for the complete model.
func generateAllGeometry(contributionsPerYear [][][]types.ContributionDay, maxContrib int, baseWidth, baseDepth float64, username, yearStr string) ([][3]float32, [][3]uint8, []uint32) {
	var positions [][3]float32
	var colors [][3]uint8
	var indices []uint32

	// Generate base with dark gray color
	basePositions, baseIndices := generateBox(0, 0, -geometry.BaseHeight, baseWidth, baseDepth, geometry.BaseHeight, uint32(len(positions)))
	for range basePositions {
		colors = append(colors, BaseColor.ToSlice())
	}
	positions = append(positions, basePositions...)
	indices = append(indices, baseIndices...)

	// Generate contribution columns for each year
	for i := len(contributionsPerYear) - 1; i >= 0; i-- {
		yearOffset := len(contributionsPerYear) - 1 - i
		baseYOffset := 2*geometry.CellSize + float64(yearOffset)*7*geometry.CellSize

		for weekIdx, week := range contributionsPerYear[i] {
			for dayIdx, day := range week {
				if day.ContributionCount > 0 {
					height := geometry.NormalizeContribution(day.ContributionCount, maxContrib)
					x := 2*geometry.CellSize + float64(weekIdx)*geometry.CellSize
					y := baseYOffset + float64(dayIdx)*geometry.CellSize

					// Get color for this contribution level
					color := ContributionToColor(day.ContributionCount, maxContrib)

					// Generate column geometry
					colPositions, colIndices := generateBox(x, y, 0, geometry.CellSize, geometry.CellSize, height, uint32(len(positions)))

					// Add colors for all vertices of this column
					for range colPositions {
						colors = append(colors, color.ToSlice())
					}

					positions = append(positions, colPositions...)
					indices = append(indices, colIndices...)
				}
			}
		}
	}

	// Generate text (username and year) - light gray for dark background
	textTriangles, err := geometry.Create3DText(username, yearStr, baseWidth, geometry.BaseHeight)
	if err == nil && len(textTriangles) > 0 {
		textPositions, textIndices := trianglesToMesh(textTriangles, uint32(len(positions)))
		for range textPositions {
			colors = append(colors, TextColor.ToSlice())
		}
		positions = append(positions, textPositions...)
		indices = append(indices, textIndices...)
	}

	// Generate logo - light gray for dark background
	logoTriangles, err := geometry.GenerateImageGeometry(baseWidth, geometry.BaseHeight)
	if err == nil && len(logoTriangles) > 0 {
		logoPositions, logoIndices := trianglesToMesh(logoTriangles, uint32(len(positions)))
		for range logoPositions {
			colors = append(colors, TextColor.ToSlice())
		}
		positions = append(positions, logoPositions...)
		indices = append(indices, logoIndices...)
	}

	return positions, colors, indices
}

// trianglesToMesh converts a slice of triangles to positions and indices for GLB.
func trianglesToMesh(triangles []types.Triangle, indexOffset uint32) ([][3]float32, []uint32) {
	positions := make([][3]float32, 0, len(triangles)*3)
	indices := make([]uint32, 0, len(triangles)*3)

	for i, tri := range triangles {
		// Add three vertices for each triangle
		positions = append(positions,
			[3]float32{float32(tri.V1.X), float32(tri.V1.Y), float32(tri.V1.Z)},
			[3]float32{float32(tri.V2.X), float32(tri.V2.Y), float32(tri.V2.Z)},
			[3]float32{float32(tri.V3.X), float32(tri.V3.Y), float32(tri.V3.Z)},
		)

		// Add indices
		base := indexOffset + uint32(i*3)
		indices = append(indices, base, base+1, base+2)
	}

	return positions, indices
}

// generateBox creates vertex positions and indices for a box.
// Returns positions as [3]float32 and indices offset by indexOffset.
func generateBox(x, y, z, width, depth, height float64, indexOffset uint32) ([][3]float32, []uint32) {
	// 8 unique vertices for the box
	// But for proper rendering with flat shading, we need 24 vertices (4 per face)
	// This allows each face to have its own vertex colors if needed

	// Define the 8 corner positions
	corners := [8][3]float32{
		{float32(x), float32(y), float32(z)},                               // 0: front-bottom-left
		{float32(x + width), float32(y), float32(z)},                       // 1: front-bottom-right
		{float32(x + width), float32(y + depth), float32(z)},               // 2: back-bottom-right
		{float32(x), float32(y + depth), float32(z)},                       // 3: back-bottom-left
		{float32(x), float32(y), float32(z + height)},                      // 4: front-top-left
		{float32(x + width), float32(y), float32(z + height)},              // 5: front-top-right
		{float32(x + width), float32(y + depth), float32(z + height)},      // 6: back-top-right
		{float32(x), float32(y + depth), float32(z + height)},              // 7: back-top-left
	}

	// For simplicity, we use 24 vertices (4 per face) to allow flat shading
	// Each face has its own set of vertices
	positions := make([][3]float32, 24)

	// Front face (y = y, facing -Y)
	positions[0], positions[1], positions[2], positions[3] = corners[0], corners[1], corners[5], corners[4]
	// Back face (y = y + depth, facing +Y)
	positions[4], positions[5], positions[6], positions[7] = corners[2], corners[3], corners[7], corners[6]
	// Left face (x = x, facing -X)
	positions[8], positions[9], positions[10], positions[11] = corners[3], corners[0], corners[4], corners[7]
	// Right face (x = x + width, facing +X)
	positions[12], positions[13], positions[14], positions[15] = corners[1], corners[2], corners[6], corners[5]
	// Top face (z = z + height, facing +Z)
	positions[16], positions[17], positions[18], positions[19] = corners[4], corners[5], corners[6], corners[7]
	// Bottom face (z = z, facing -Z)
	positions[20], positions[21], positions[22], positions[23] = corners[3], corners[2], corners[1], corners[0]

	// Indices for 6 faces, 2 triangles each, counter-clockwise winding
	baseIndices := []uint32{
		// Front
		0, 1, 2, 0, 2, 3,
		// Back
		4, 5, 6, 4, 6, 7,
		// Left
		8, 9, 10, 8, 10, 11,
		// Right
		12, 13, 14, 12, 14, 15,
		// Top
		16, 17, 18, 16, 18, 19,
		// Bottom
		20, 21, 22, 20, 22, 23,
	}

	// Apply offset to indices
	indices := make([]uint32, len(baseIndices))
	for i, idx := range baseIndices {
		indices[i] = idx + indexOffset
	}

	return positions, indices
}

// findMaxContributionsAcrossYears finds the maximum contribution count across all years.
func findMaxContributionsAcrossYears(contributionsPerYear [][][]types.ContributionDay) int {
	maxContrib := 0
	for _, yearContributions := range contributionsPerYear {
		for _, week := range yearContributions {
			for _, day := range week {
				if day.ContributionCount > maxContrib {
					maxContrib = day.ContributionCount
				}
			}
		}
	}
	return maxContrib
}
