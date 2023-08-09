const query = @import("query.zig");

pub fn QueryPrinter(comptime T: type) type {
    return struct {
        const Self = @This();

        w: T,

        pub fn init(w: T) Self {
            return Self { .w = w };
        }

        pub fn print(self: *Self, item: query.Query) void {
            switch (item) {
                .select => self.printSelectQuery(item.select),
                .insert => self.printInsertQuery(item.insert),
                .err => unreachable, // TODO
            }
        }

        fn printSelectQuery(self: *Self, item: query.SelectQuery) void {
            self.printSelectClause(item.select);

            if (item.from) |acc| {
                self.printFromClause(acc);
            }

            if (item.where) |acc| {
                self.w.print(" ", .{}) catch return;
                self.printWhereClause(acc);
            }

            if (item.having) |acc| {
                self.w.print(" ", .{}) catch return;
                self.printHavingClause(acc);
            }

            if (item.group_by) |acc| {
                self.w.print(" ", .{}) catch return;
                self.printGroupByClause(acc);
            }

            if (item.order_by) |acc| {
                self.w.print(" ", .{}) catch return;
                self.printOrderByClause(acc);
            }

            if (item.limit) |acc| {
                self.w.print(" ", .{}) catch return;
                self.printLimitClause(acc);
            }
        }

        fn printInsertQuery(_: *Self, _: query.InsertQuery) void {
            unreachable;
        }

        fn printSelectClause(self: *Self, item: query.SelectClause) void {
            self.w.print("SELECT ", .{}) catch return;

            // if (item.distinct) {
            //     self.w.print("DISTINCT ", .{});
            // }

            for (item.column_list) |column| {
                self.printColumnExpression(column);
            }
        }

        fn printFromClause(self: *Self, item: query.FromClause) void {
            self.w.print("FROM {s}", .{item.tableName}) catch return;
        }

        fn printWhereClause(self: *Self, item: query.WhereClause) void {
            self.w.print("WHERE ", .{}) catch return;
            self.printExpression(item.predicate);
        }

        fn printHavingClause(self: *Self, item: query.HavingClause) void {
            self.w.print("HAVING ", .{}) catch return;
            self.printExpression(item.predicate);
        }

        fn printGroupByClause(self: *Self, item: query.GroupByClause) void {
            self.w.print("GROUP BY ", .{}) catch return;
            _ = item;
            // TODO
            unreachable;
        }

        fn printOrderByClause(self: *Self, item: query.OrderByClause) void {
            self.w.print("ORDER BY ", .{}) catch return;
            _ = item;
            // TODO
            unreachable;
        }

        fn printLimitClause(self: *Self, item: query.LimitClause) void {
            self.w.print("LIMIT {d}", .{item.limit}) catch return;
            if (item.offset) |offset| {
                self.w.print(" OFFSET {d}", .{offset}) catch return;
            }
        }

        fn printColumnExpression(self: *Self, item: query.ColumnExpression) void {
            self.printExpression(item.expr);
            if (item.alias) |alias| {
                self.w.print(" AS {s}", .{alias}) catch return;
            }
        }

        fn printExpression(self: *Self, item: query.Expression) void {
            switch (item) {
                .string => self.w.print("\"{s}\"", .{ item.string }) catch return, // TODO: Escape string.
                .integer => self.w.print("{d}", .{ item.integer }) catch return,
                .floating_point => self.w.print("{f}", .{ item.floating_point }) catch return,
                .boolean => self.w.print("{s}", .{ if (item.boolean) "TRUE" else "FALSE" }) catch return,
                .identifier => self.w.print("{s}", .{ item.identifier }) catch return,
                .equal => {
                    self.printExpression(item.equal.left.*);
                    self.w.print(" = ", .{}) catch return;
                    self.printExpression(item.equal.right.*);
                },
                .not_equal => {
                    self.printExpression(item.not_equal.left.*);
                    self.w.print(" != ", .{}) catch return;
                    self.printExpression(item.not_equal.right.*);
                },
                .less_than => {
                    self.printExpression(item.less_than.left.*);
                    self.w.print(" < ", .{}) catch return;
                    self.printExpression(item.less_than.right.*);
                },
                .less_than_or_equal => {
                    self.printExpression(item.less_than_or_equal.left.*);
                    self.w.print(" <= ", .{}) catch return;
                    self.printExpression(item.less_than_or_equal.right.*);
                },
                .greater_than => {
                    self.printExpression(item.greater_than.left.*);
                    self.w.print(" > ", .{}) catch return;
                    self.printExpression(item.greater_than.right.*);
                },
                .greater_than_or_equal => {
                    self.printExpression(item.greater_than_or_equal.left.*);
                    self.w.print(" >= ", .{}) catch return;
                    self.printExpression(item.greater_than_or_equal.right.*);
                },
                .bool_and => {
                    self.w.print("(", .{}) catch return;
                    self.printExpression(item.bool_and.left.*);
                    self.w.print(" AND ", .{}) catch return;
                    self.printExpression(item.bool_and.right.*);
                    self.w.print(")", .{}) catch return;
                },
                .bool_or => {
                    self.w.print("(", .{}) catch return;
                    self.printExpression(item.bool_or.left.*);
                    self.w.print(" OR ", .{}) catch return;
                    self.printExpression(item.bool_or.right.*);
                    self.w.print(")", .{}) catch return;
                },
                .bool_not => {
                    self.w.print("(NOT ", .{}) catch return;
                    self.printExpression(item.bool_not.*);
                    self.w.print(")", .{}) catch return;
                },
                else => {
                    self.w.print("<unimplemented>", .{}) catch return;
                }
            }
        }

        fn printBinaryExpression(_: *Self, _: query.BinaryExpression) void {
            unreachable;
        }

        fn printFunctionCall(self: *Self, item: query.FunctionCall) void {
            self.w.print("{s}(", .{ item.name }) catch return;
            for (item.args) |arg| {
                self.printExpression(arg);
            }
            self.w.print(")", .{}) catch return;
        }
    };
}
