package main

import (
	"fmt"
	"testing"
)

func TestConv(t *testing.T) {
	tests := []FloatCoords{
		{85.05, -180},
		{90, -180}, // upper left
		{90, 180},  // upper right
		{0, -180},
		{0, 0},
		{0, 180},
		{-90, -180},
		{-90, 180},
		{49.865622, 8.6497177},
	}
	for _, c := range tests {
		fmt.Println(c, c.Transform())
	}
}
