// Package terrain paints a mission map: ground, water, bridges, ramps,
// mines, and tree canopies. The battle and the map editor both call Prepare.
//
// Read this file first. Paint is terrain_paint.go. Plates are plates/*.png.
// water.kage shimmers water, sand, and ice.
//
// Scale is offscreen pixels per world pixel. Time is seconds. The caller
// passes both. DrawGround does not draw trees. Canopies stay separate so
// the battle can Y-sort them with huts and soldiers.
//
// The sim import is only Map, Tile, and TileSize. This package must not
// import internal/render, internal/app, internal/audio, internal/data,
// or internal/campaign.
package terrain
