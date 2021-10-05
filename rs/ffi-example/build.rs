extern crate bindgen;

use std::{env, path::PathBuf};

fn main() {
    println!("cargo:rerun-if-changed=bindings.h");
    println!("cargo:rustc-link-lib=dylib=point");
    let bindings = bindgen::Builder::default()
        .header("bindings.h")
        .parse_callbacks(Box::new(bindgen::CargoCallbacks))
        .generate()
        .expect("Unable to generate bindings");
    let out_path = PathBuf::from(env::var("OUT_DIR").unwrap());
    bindings
        .write_to_file(out_path.join("bindings.rs"))
        .expect("Couldn't write bindings!");
}
