pub const Header = packed struct {
    next: ?*Header = null,
    freeStart: u16,
    freeLen: u16,
};


