import glyph_raster;

fn main() -> i32 effects { mut } {
    let ops: [i32; 5] = [0, 1, 1, 1, 1];
    let z: f32 = 0 as f32;
    let hundred: f32 = 100 as f32;
    let pts: [f32; 10] = [z, z, hundred, z, hundred, hundred, z, hundred, z, z];
    let edges: [f32; 32] = [
        z, z, z, z, z, z, z, z, z, z, z, z, z, z, z, z,
        z, z, z, z, z, z, z, z, z, z, z, z, z, z, z, z
    ];
    let xs: [f32; 16] = [z, z, z, z, z, z, z, z, z, z, z, z, z, z, z, z];
    let dirs: [i32; 16] = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0];
    let rgba: [u8; 64] = [
        0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8,
        0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8,
        0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8,
        0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8,
        0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8,
        0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8,
        0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8,
        0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8, 0 as u8
    ];
    let scale: f32 = (3 as f32) / (100 as f32);
    let ox: f32 = (1 as f32) / (2 as f32);
    let oy: f32 = (7 as f32) / (2 as f32);
    let ink: i32 = glyph_raster.raster_outline(&ops, &pts, 4, 4, scale, ox, oy, &mut edges, &mut xs, &mut dirs, &mut rgba);
    if (ink < 1) {
        return 1;
    }
    return 42;
}