// Package terrain paints a mission map: ground, water, bridges, ramps,
// mines, and tree canopies. A region phase is PrepareChart. The battle
// calls PrepareChartLive so a spent mine drops its mark. A tile phase,
// including the sandboxes, calls Prepare.
//
// Read this file first. Paint is terrain_paint.go. Each terrain has its own
// canopy, shallow, and deep plates under plates/<kind>/. Jungle open ground
// is grassColor, not plates/jungle/ground.png. Snow open ground is
// snowColor, not plates/snow/ground.png. Desert open ground is
// sandColor, not plates/desert/ground.png. Forest open ground is
// forestColor, not plates/forest/ground.png. Water fades from the
// shore: pale near the sand, deep further out, and the line waves
// instead of following the cell edge. Shallow and deep plates stay
// on disk and are not sampled for that fade. A snow cliff's edge
// waves a few pixels, like the shore, and the face is the cliff
// palette rather than the rock plate. Rock, ice, dirt, and
// quicksand stay in plates/. water.kage shimmers water, sand, and ice.
//
// Scale is offscreen pixels per world pixel. Time is seconds. The caller
// passes both. DrawGround does not draw trees. Canopies stay separate so
// the battle can Y-sort them with huts and soldiers.
//
// The sim import is only Map, Tile, and TileSize. This package must not
// import internal/render, internal/app, internal/audio, internal/data,
// or internal/campaign.
package terrain
