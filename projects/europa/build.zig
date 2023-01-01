const std = @import("std");

pub fn build(b: *std.build.Builder) void {
    const lib = b.addSharedLibrary("xyzzy", "src/main.zig", .unversioned);
    // const lib = b.addStaticLibrary("xyzzy", "src/main.zig");

    // Standard release options allow the person running `zig build` to select
    // between Debug, ReleaseSafe, ReleaseFast, and ReleaseSmall.
    const mode = b.standardReleaseOptions();
    lib.setBuildMode(mode);

    lib.setTarget(.{
      .cpu_arch = .wasm32,
      .os_tag = .freestanding,
    });

    lib.setOutputDir("public/assets/wasm");
    b.default_step.dependOn(&lib.step);

    var main_tests = b.addTest("src/main.zig");
    main_tests.setBuildMode(mode);

    const test_step = b.step("test", "Run library tests");
    test_step.dependOn(&main_tests.step);
}
