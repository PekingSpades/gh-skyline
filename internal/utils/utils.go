// Package utils are utility functions for the GitHub Skyline project
package utils //nolint:revive // package name is appropriate for this internal module

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Constants for GitHub launch year and default output file format
const (
	githubLaunchYear   = 2008
	outputFileFormatSTL = "%s-%s-github-skyline.stl"
	outputFileFormatGLB = "%s-%s-github-skyline.glb"
)

// ParseYearRange parses whether a year is a single year or a range of years.
func ParseYearRange(yearRange string) (startYear, endYear int, err error) {
	if strings.Contains(yearRange, "-") {
		parts := strings.Split(yearRange, "-")
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("invalid year range format")
		}
		startYear, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, err
		}
		endYear, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, err
		}
	} else {
		year, err := strconv.Atoi(yearRange)
		if err != nil {
			return 0, 0, err
		}
		startYear, endYear = year, year
	}
	return startYear, endYear, validateYearRange(startYear, endYear)
}

// validateYearRange checks if the years are within the range
// of GitHub's launch year to the current year and if
// the start year is not greater than the end year.
func validateYearRange(startYear, endYear int) error {
	currentYear := time.Now().Year()
	if startYear < githubLaunchYear || endYear > currentYear {
		return fmt.Errorf("years must be between %d and %d", githubLaunchYear, currentYear)
	}
	if startYear > endYear {
		return fmt.Errorf("start year cannot be after end year")
	}
	return nil
}

// FormatYearRange returns a formatted string representation of the year range
func FormatYearRange(startYear, endYear int) string {
	if startYear == endYear {
		return fmt.Sprintf("%d", startYear)
	}
	// Use YYYY-YY format for multi-year ranges
	return fmt.Sprintf("%04d-%02d", startYear, endYear%100)
}

// GenerateOutputFilename creates a consistent filename for the output file
func GenerateOutputFilename(user string, startYear, endYear int, output string, format string) string {
	ext := ".stl"
	outputFormat := outputFileFormatSTL
	if strings.ToLower(format) == "glb" {
		ext = ".glb"
		outputFormat = outputFileFormatGLB
	}

	if output != "" {
		// Ensure the filename ends with the correct extension
		lowerOutput := strings.ToLower(output)
		if !strings.HasSuffix(lowerOutput, ".stl") && !strings.HasSuffix(lowerOutput, ".glb") {
			return output + ext
		}
		return output
	}
	yearStr := FormatYearRange(startYear, endYear)
	return fmt.Sprintf(outputFormat, user, yearStr)
}
