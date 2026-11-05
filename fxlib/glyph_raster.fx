// Gru Phase 2 — outline → RGBA8 glyph raster (library; no main).
// Symbols: fx_glyph_raster_*
// ops: 0=Move 1=Line 2=Quad 3=Cube; pts: flat x,y font units.
// edges/xs/dirs: scratch; rgba: preallocated w*h*4.
// Fill: non-zero winding (TrueType / gg.Fill parity) + 4× coverage AA
// (not LCD subpixel — see T1_TEXT_SKIA_BORROWINGS.md).
// Note: use `(n as f32)` not `n.0f32` — emit-C FXC-1 (see docs/FX_CONSUMER_ISSUES.md).
//
// boost_rgba_alpha — Skia-lite contrast for greyscale glyph atlases (T1.15 / no-SSAA).
// Raises mid-tone coverage so dark-on-light text reads less washed at SSAA 1×.

fn clear_rgba(rgba: &mut [u8]) -> i32 effects { mut } {
    let n: i32 = rgba.len;
    for (let i: i32 = 0; i < n; i = i + 1) {
        rgba[i] = 0 as u8;
    }
    return n;
}

fn px_x(x: f32, scale: f32, ox: f32) -> f32 {
    return x * scale + ox;
}

fn px_y(y: f32, scale: f32, oy: f32) -> f32 {
    return oy - y * scale;
}

fn push_edge(edges: &mut [f32], n: i32, x0: f32, y0: f32, x1: f32, y1: f32) -> i32 effects { mut } {
    if (n + 4 > edges.len) {
        return -2;
    }
    if (y0 == y1 && x0 == x1) {
        return n;
    }
    edges[n] = x0;
    edges[n + 1] = y0;
    edges[n + 2] = x1;
    edges[n + 3] = y1;
    return n + 4;
}

fn quad_eval(p0: f32, p1: f32, p2: f32, t: f32) -> f32 {
    let one: f32 = 1 as f32;
    let two: f32 = 2 as f32;
    let u: f32 = one - t;
    let r: f32 = u * u * p0 + two * u * t * p1 + t * t * p2;
    return r;
}

fn cube_eval(p0: f32, p1: f32, p2: f32, p3: f32, t: f32) -> f32 {
    let one: f32 = 1 as f32;
    let three: f32 = 3 as f32;
    let u: f32 = one - t;
    let uu: f32 = u * u;
    let tt: f32 = t * t;
    let r: f32 = uu * u * p0 + three * uu * t * p1 + three * u * tt * p2 + tt * t * p3;
    return r;
}

fn flatten_outline(ops: &[i32], pts: &[f32], scale: f32, ox: f32, oy: f32, edges: &mut [f32]) -> i32 effects { mut } {
    let n: i32 = 0;
    let z: f32 = 0 as f32;
    let cx: f32 = z;
    let cy: f32 = z;
    let pi: i32 = 0;
    let steps: i32 = 24;
    let nops: i32 = ops.len;
    for (let oi: i32 = 0; oi < nops; oi = oi + 1) {
        let op: i32 = ops[oi];
        if (op == 0) {
            if (pi + 2 > pts.len) {
                return -1;
            }
            cx = px_x(pts[pi], scale, ox);
            cy = px_y(pts[pi + 1], scale, oy);
            pi = pi + 2;
        } else {
            if (op == 1) {
                if (pi + 2 > pts.len) {
                    return -1;
                }
                let x1: f32 = px_x(pts[pi], scale, ox);
                let y1: f32 = px_y(pts[pi + 1], scale, oy);
                pi = pi + 2;
                n = push_edge(edges, n, cx, cy, x1, y1);
                if (n < 0) {
                    return n;
                }
                cx = x1;
                cy = y1;
            } else {
                if (op == 2) {
                    if (pi + 4 > pts.len) {
                        return -1;
                    }
                    let c1x: f32 = px_x(pts[pi], scale, ox);
                    let c1y: f32 = px_y(pts[pi + 1], scale, oy);
                    let x2: f32 = px_x(pts[pi + 2], scale, ox);
                    let y2: f32 = px_y(pts[pi + 3], scale, oy);
                    pi = pi + 4;
                    let x0: f32 = cx;
                    let y0: f32 = cy;
                    for (let s: i32 = 0; s < steps; s = s + 1) {
                        let t0: f32 = (s as f32) / (steps as f32);
                        let t1: f32 = ((s + 1) as f32) / (steps as f32);
                        let ax: f32 = quad_eval(x0, c1x, x2, t0);
                        let ay: f32 = quad_eval(y0, c1y, y2, t0);
                        let bx: f32 = quad_eval(x0, c1x, x2, t1);
                        let by: f32 = quad_eval(y0, c1y, y2, t1);
                        n = push_edge(edges, n, ax, ay, bx, by);
                        if (n < 0) {
                            return n;
                        }
                    }
                    cx = x2;
                    cy = y2;
                } else {
                    if (op == 3) {
                        if (pi + 6 > pts.len) {
                            return -1;
                        }
                        let c1x: f32 = px_x(pts[pi], scale, ox);
                        let c1y: f32 = px_y(pts[pi + 1], scale, oy);
                        let c2x: f32 = px_x(pts[pi + 2], scale, ox);
                        let c2y: f32 = px_y(pts[pi + 3], scale, oy);
                        let x3: f32 = px_x(pts[pi + 4], scale, ox);
                        let y3: f32 = px_y(pts[pi + 5], scale, oy);
                        pi = pi + 6;
                        let x0: f32 = cx;
                        let y0: f32 = cy;
                        for (let s: i32 = 0; s < steps; s = s + 1) {
                            let t0: f32 = (s as f32) / (steps as f32);
                            let t1: f32 = ((s + 1) as f32) / (steps as f32);
                            let ax: f32 = cube_eval(x0, c1x, c2x, x3, t0);
                            let ay: f32 = cube_eval(y0, c1y, c2y, y3, t0);
                            let bx: f32 = cube_eval(x0, c1x, c2x, x3, t1);
                            let by: f32 = cube_eval(y0, c1y, c2y, y3, t1);
                            n = push_edge(edges, n, ax, ay, bx, by);
                            if (n < 0) {
                                return n;
                            }
                        }
                        cx = x3;
                        cy = y3;
                    } else {
                        return -1;
                    }
                }
            }
        }
    }
    return n;
}

// Accumulate greyscale coverage into rgba alpha (RGB set to white).
fn add_cover(rgba: &mut [u8], w: i32, h: i32, x: i32, y: i32, add: i32) -> i32 effects { mut } {
    if (add <= 0) {
        return 0;
    }
    if (x < 0) {
        return 0;
    }
    if (y < 0) {
        return 0;
    }
    if (x >= w) {
        return 0;
    }
    if (y >= h) {
        return 0;
    }
    let o: i32 = (y * w + x) * 4;
    let v: i32 = (rgba[o + 3] as i32) + add;
    if (v > 255) {
        v = 255;
    }
    rgba[o] = 255 as u8;
    rgba[o + 1] = 255 as u8;
    rgba[o + 2] = 255 as u8;
    rgba[o + 3] = v as u8;
    return 1;
}

// Horizontal coverage AA for one sub-scan span [x0f, x1f) with weight (e.g. 64×4≈255).
fn fill_span_aa(rgba: &mut [u8], w: i32, h: i32, y: i32, x0f: f32, x1f: f32, weight: i32) -> i32 effects { mut } {
    if (x1f <= x0f) {
        return 0;
    }
    let z: f32 = 0 as f32;
    let one: f32 = 1 as f32;
    let xf0: f32 = x0f;
    let xf1: f32 = x1f;
    if (xf0 < z) {
        xf0 = z;
    }
    if (xf1 > (w as f32)) {
        xf1 = w as f32;
    }
    if (xf1 <= xf0) {
        return 0;
    }
    let ink: i32 = 0;
    let x: i32 = xf0 as i32;
    while ((x as f32) < xf1) {
        if (x >= w) {
            break;
        }
        if (x >= 0) {
            let left: f32 = x as f32;
            let right: f32 = left + one;
            if (xf0 > left) {
                left = xf0;
            }
            if (xf1 < right) {
                right = xf1;
            }
            let cover: f32 = right - left;
            if (cover > z) {
                let add: i32 = ((cover * (weight as f32)) as i32);
                ink = ink + add_cover(rgba, w, h, x, y, add);
            }
        }
        x = x + 1;
    }
    return ink;
}

// Non-zero winding fill with 4× vertical + fractional horizontal coverage AA
// (closes binary-edge choppiness vs gg; not LCD/subpixel — that stays deferred).
fn fill_nonzero(edges: &mut [f32], edge_n: i32, w: i32, h: i32, xs: &mut [f32], dirs: &mut [i32], rgba: &mut [u8]) -> i32 effects { mut } {
    let need: i32 = w * h * 4;
    if (rgba.len < need) {
        return -3;
    }
    if (xs.len < 2) {
        return -1;
    }
    if (dirs.len < xs.len) {
        return -1;
    }
    let ink: i32 = 0;
    let xmax: i32 = xs.len;
    let samples: i32 = 4;
    let weight: i32 = 64;
    let four: f32 = 4 as f32;
    let half_sub: f32 = (1 as f32) / (8 as f32);
    for (let y: i32 = 0; y < h; y = y + 1) {
        for (let s: i32 = 0; s < samples; s = s + 1) {
            let ymid: f32 = (y as f32) + ((s as f32) / four) + half_sub;
            let xc: i32 = 0;
            let e: i32 = 0;
            while (e + 3 < edge_n) {
                let x0: f32 = edges[e];
                let y0: f32 = edges[e + 1];
                let x1: f32 = edges[e + 2];
                let y1: f32 = edges[e + 3];
                e = e + 4;
                let ymin: f32 = y0;
                let ymax: f32 = y1;
                let dir: i32 = 1;
                if (y1 < y0) {
                    ymin = y1;
                    ymax = y0;
                    dir = -1;
                }
                if (ymid >= ymin) {
                    if (ymid < ymax) {
                        let dy: f32 = y1 - y0;
                        let t: f32 = (ymid - y0) / dy;
                        let xi: f32 = x0 + (x1 - x0) * t;
                        if (xc < xmax) {
                            xs[xc] = xi;
                            dirs[xc] = dir;
                            xc = xc + 1;
                        }
                    }
                }
            }
            for (let i: i32 = 1; i < xc; i = i + 1) {
                let key: f32 = xs[i];
                let keyd: i32 = dirs[i];
                let j: i32 = i - 1;
                while (j >= 0) {
                    if (xs[j] <= key) {
                        break;
                    }
                    xs[j + 1] = xs[j];
                    dirs[j + 1] = dirs[j];
                    j = j - 1;
                }
                xs[j + 1] = key;
                dirs[j + 1] = keyd;
            }
            let wind: i32 = 0;
            let p: i32 = 0;
            while (p + 1 < xc) {
                wind = wind + dirs[p];
                if (wind != 0) {
                    ink = ink + fill_span_aa(rgba, w, h, y, xs[p], xs[p + 1], weight);
                }
                p = p + 1;
            }
        }
    }
    return ink;
}

fn raster_outline(ops: &[i32], pts: &[f32], w: i32, h: i32, scale: f32, ox: f32, oy: f32, edges: &mut [f32], xs: &mut [f32], dirs: &mut [i32], rgba: &mut [u8]) -> i32 effects { mut } {
    if (w <= 0) {
        return -1;
    }
    if (h <= 0) {
        return -1;
    }
    let _clr: i32 = clear_rgba(rgba);
    let edge_n: i32 = flatten_outline(ops, pts, scale, ox, oy, edges);
    if (edge_n < 0) {
        return edge_n;
    }
    return fill_nonzero(edges, edge_n, w, h, xs, dirs, rgba);
}

// boost_rgba_alpha — Skia-lite contrast on RGBA8 alpha (stride 4).
// coverage' = a + contrast*a*(1-a). contrast in 0..1. Returns pixels touched.
fn boost_rgba_alpha(rgba: &mut [u8], contrast: f32) -> i32 effects { mut } {
    let n: i32 = rgba.len;
    if (n < 4) {
        return -1;
    }
    let zero: f32 = 0 as f32;
    let one: f32 = 1 as f32;
    let c: f32 = contrast;
    if (c <= zero) {
        return 0;
    }
    let touched: i32 = 0;
    let i: i32 = 3;
    while (i < n) {
        let a: f32 = (rgba[i] as f32) / (255 as f32);
        if (a > zero) {
            if (a < one) {
                let boosted: f32 = a + c * a * (one - a);
                let v: i32 = (boosted * (255 as f32) + (1 as f32) / (2 as f32)) as i32;
                if (v > 255) {
                    rgba[i] = 255 as u8;
                } else {
                    if (v < 0) {
                        rgba[i] = 0 as u8;
                    } else {
                        rgba[i] = v as u8;
                    }
                }
                touched = touched + 1;
            }
        }
        i = i + 4;
    }
    return touched;
}