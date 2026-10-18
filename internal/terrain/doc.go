// Package terrain paints a mission map: ground, water, bridges, ramps,
// mines, and tree canopies. A region phase is PrepareChart. The battle
// calls PrepareChartLive so a spent mine drops its mark. A tile phase,
// including the sandboxes, calls Prepare.
//
// Read this file first. Paint is terrain_paint.go. Each terrain has its own
// canopy, shallow, and deep plates under plates/<kind>/. Jungle open ground
// is grassColor, not plates/jungle/ground.png. Snow open ground is
// snowColor, not plates/snow/ground.png. Rock, ice,
// dirt, and quicksand stay in plates/. water.kage shimmers water, sand, and ice.
//
// Scale is offscreen pixels per world pixel. Time is seconds. The caller
// passes both. DrawGround does not draw trees. Canopies stay separate so
// the battle can Y-sort them with huts and soldiers.
//
// The sim import is only Map, Tile, and TileSize. This package must not
// import internal/render, internal/app, internal/audio, internal/data,
// or internal/campaign.
package terrain
