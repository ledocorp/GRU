import measure_wrap;

fn main() -> i32 effects { mut } {
    let adv: [f32; 8] = [
        10 as f32, 10 as f32, 5 as f32, 10 as f32,
        10 as f32, 5 as f32, 10 as f32, 10 as f32
    ];
    let brk: [i32; 8] = [0, 0, 1, 0, 0, 1, 0, 0];
    let out: [i32; 8] = [0, 0, 0, 0, 0, 0, 0, 0];
    let n: i32 = measure_wrap.wrap_advances(&adv, &brk, 25 as f32, &mut out);
    if (n != 3) {
        return 1;
    }
    if (out[0] != 3) {
        return 2;
    }
    if (out[1] != 6) {
        return 3;
    }
    if (out[2] != 8) {
        return 4;
    }
    let s: f32 = measure_wrap.sum_advances(&adv, 0, 3);
    if (s != 25 as f32) {
        return 5;
    }
    return 42;
}