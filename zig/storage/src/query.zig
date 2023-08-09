const std = @import("std");
const String = @import("util.zig").String;

// This file could be generated from a grammar file of some sort.

pub const Query = union(enum) {
    select: SelectQuery,
    insert: InsertQuery,
    err: String,
};

pub const SelectQuery = struct {
    const Self = @This();

    select: SelectClause,
    from: ?FromClause,
    where: ?WhereClause,
    group_by: ?GroupByClause,
    having: ?HavingClause,
    order_by: ?OrderByClause,
    limit: ?LimitClause,
};

pub const SelectClause = struct {
    const Self = @This();

    column_list: []ColumnExpression,
};

pub const FromClause = struct {
    const Self = @This();

    tableName: String,
};

pub const WhereClause = struct {
    const Self = @This();

    predicate: Expression,
};

pub const HavingClause = struct {
    const Self = @This();

    predicate: Expression,
};

pub const GroupByClause = struct {
    const Self = @This();

    column_list: []ColumnExpression,
};

pub const OrderByClause = struct {
    const Self = @This();

    column_list: []ColumnExpression,
};

pub const LimitClause = struct {
    const Self = @This();

    limit: u64,
    offset: ?u64,
};

// END SELECT

//////////////////////
/// Everything below this point is not fleshed out.

// BEGIN INSERT

pub const InsertQuery = struct {
    const Self = @This();
};

// BEGIN COMMON TYPES

pub const ColumnExpression = struct {
    const Self = @This();

    expr: Expression,
    alias: ?String,
};

pub const Expression = union(enum) {
    const Self = @This();

    string: String,
    integer: i64,
    floating_point: f64,
    boolean: bool,
    identifier: String,

    equal: BinaryExpression,
    not_equal: BinaryExpression,
    greater_than: BinaryExpression,
    greater_than_or_equal: BinaryExpression,
    less_than: BinaryExpression,
    less_than_or_equal: BinaryExpression,
    bool_and: BinaryExpression,
    bool_or: BinaryExpression,
    add: BinaryExpression,
    subtract: BinaryExpression,
    multiply: BinaryExpression,
    divide: BinaryExpression,
    modulo: BinaryExpression,
    bool_not: *Expression,
    unary_minus: *Expression,
    function_call: FunctionCall,
};

pub const BinaryExpression = struct {
    const Self = @This();

    left: *Expression,
    right: *Expression,
};

pub const FunctionCall = struct {
    const Self = @This();

    name: String,
    args: []Expression,
};
