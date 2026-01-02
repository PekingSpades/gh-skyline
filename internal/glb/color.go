// Package glb provides functionality for generating GLB (binary glTF) files
// with vertex colors from GitHub contribution data.
package glb

// RGB represents an RGB color with 8-bit components.
type RGB struct {
	R, G, B uint8
}

// GitHubContributionColors defines the GitHub Dark Mode contribution graph colors.
// These colors are optimized for display on dark backgrounds.
var GitHubContributionColors = []RGB{
	{22, 27, 34},  // Level 0: #161b22 - no contributions (dark gray)
	{14, 68, 41},  // Level 1: #0e4429 - deep green
	{0, 109, 50},  // Level 2: #006d32 - medium green
	{38, 166, 65}, // Level 3: #26a641 - bright green
	{57, 211, 83}, // Level 4: #39d353 - neon green
}

// BaseColor is the color for the model base (dark gray).
var BaseColor = RGB{33, 38, 45} // #21262d

// TextColor is the color for text and logo (light gray).
var TextColor = RGB{201, 209, 217} // #c9d1d9

// ContributionToColor maps a contribution count to a GitHub Dark Mode color.
// The color intensity is based on the ratio of count to maxCount.
func ContributionToColor(count, maxCount int) RGB {
	if count == 0 || maxCount == 0 {
		return GitHubContributionColors[0]
	}

	ratio := float64(count) / float64(maxCount)

	switch {
	case ratio < 0.25:
		return GitHubContributionColors[1]
	case ratio < 0.50:
		return GitHubContributionColors[2]
	case ratio < 0.75:
		return GitHubContributionColors[3]
	default:
		return GitHubContributionColors[4]
	}
}

// ToSlice converts RGB to a [3]uint8 slice for use with the gltf library.
func (c RGB) ToSlice() [3]uint8 {
	return [3]uint8{c.R, c.G, c.B}
}
