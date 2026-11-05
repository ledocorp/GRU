// Gru Phase 3 — wrap pre-shaped advances (library; no main).
// Symbols: fx_measure_wrap_*
// advances[i] = glyph/cluster width; can_break[i]=1 if break allowed after i (e.g. space).
// out_ends[k] = exclusive end index of line k; returns line count (or -1).

fn wrap_advances(advances: &[f32], can_break: &[i32], max_w: f32, out_ends: &mut [i32]) -> i32 effects { mut } {
    let n: i32 = advances.len;
    if (can_break.len < n) {
        return -1;
    }
    if (out_ends.len < 1) {
        return -1;
    }
    if (n == 0) {
        return 0;
    }
    if (max_w <= 0 as f32) {
        return -1;
    }
    let lines: i32 = 0;
    let i: i32 = 0;
    while (i < n) {
        let line_w: f32 = 0 as f32;
        let start: i32 = i;
        let last_break: i32 = -1;
        while (i < n) {
            let adv: f32 = advances[i];
            if (line_w + adv > max_w) {
                if (line_w > 0 as f32) {
                    break;
                }
                // Single item wider than max — force one glyph per line.
                i = i + 1;
                if (lines >= out_ends.len) {
                    return -2;
                }
                out_ends[lines] = i;
                lines = lines + 1;
                last_break = -1;
                line_w = 0 as f32;
                start = i;
                continue;
            }
            line_w = line_w + adv;
            if (can_break[i] != 0) {
                last_break = i + 1;
            }
            i = i + 1;
        }
        let cut: i32 = i;
        if (i < n) {
            if (last_break > start) {
                cut = last_break;
                i = cut;
            }
        }
        if (cut <= start) {
            cut = start + 1;
            i = cut;
        }
        if (lines >= out_ends.len) {
            return -2;
        }
        out_ends[lines] = cut;
        lines = lines + 1;
    }
    return lines;
}

fn sum_advances(advances: &[f32], lo: i32, hi: i32) -> f32 {
    let n: i32 = advances.len;
    let a: i32 = lo;
    let b: i32 = hi;
    if (a < 0) {
        a = 0;
    }
    if (b > n) {
        b = n;
    }
    let s: f32 = 0 as f32;
    let i: i32 = a;
    while (i < b) {
        s = s + advances[i];
        i = i + 1;
    }
    return s;
}