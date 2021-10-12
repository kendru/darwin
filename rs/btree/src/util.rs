
// TODO: Replace most call sites with Layout::padding_needed_for <https://github.com/rust-lang/rust/issues/55724>
// when stabilized.
pub(crate) fn pad_for(n: usize, multiple: usize) -> usize {
    round_to(n, multiple) - n
}

pub(crate) fn round_to(n: usize, multiple: usize) -> usize {
    match n % multiple {
        0 => n,
        rem => n - rem + multiple,
    }
}
