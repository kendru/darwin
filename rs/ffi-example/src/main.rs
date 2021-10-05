#![allow(non_upper_case_globals)]
#![allow(non_camel_case_types)]
#![allow(non_snake_case)]

include!(concat!(env!("OUT_DIR"), "/bindings.rs"));

fn main() {
    println!("Hello, world!");
    let p1 = &mut Point{
        x: 0,
        y: 0,
    };
    let p2 = &mut Point{
        x: 3,
        y: 4,
    };
    let dist = unsafe {
        point_distance(p1 as *mut Point, p2 as *mut Point)
    };

    println!("Distance between {:?} and {:?} is {}", p1, p2, dist);
}
