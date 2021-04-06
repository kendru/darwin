# JCrunch

## Spec

- In general, all values are stored as a type tag followed by an encoded value.
- Integers are encoded as big-endian numbers in the smallest type that can hold the number. For instance, `234` will be encoded as an unsigned 8-bit integer, and `-8435` will be encoded as a 16-bit signed integer.

#### Reading a Number

```
   INSTR         STACK
   ========      ========
-> POP_UINT8  -> 123
   ========      ========
= 123
```

#### Reading a String

```
   INSTR           STACK           DATA
   ========        ========        ========
-> LOAD 0                          0: "Hello, World"
   POP_STR
   ========        ========        ========

   INSTR           STACK           DATA
   ========        ========        ========
   LOAD 0       -> "Hello, World"  0: "Hello, World"
-> POP_STR
   ========        ========        ========
= "Hello, World"
```

#### Reading an Array

```
   INSTR            STACK           DATA
   ========         ========        ========
-> START_ARR     -> 1               0: 0     ; loop counter
   EQ_NUM    0 8    2
   JMP_F     5      3
   POP              4
   JMP       9      5
   POP_UINT8        6
   APPEND_ARR       7
   INC_NUM   0      8
   JMP       1
   END_ARR
   ========         ========        ========
Internal State: <nil>

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     -> 1               0: 0     ; loop counter
-> EQ_NUM    0 8    2
   JMP_F     5      3
   POP              4
   JMP       9      5
   POP_UINT8        6
   APPEND_ARR       7
   INC_NUM   0      8
   JMP       1
   END_ARR
   ========         ========        ========
Internal State: []

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     -> false           0: 0     ; loop counter
   EQ_NUM    0 8    1
-> JMP_F     5      2
   POP              3
   JMP       9      4
   POP_UINT8        5
   APPEND_ARR       6
   INC_NUM   0      7
   JMP       1      8
   END_ARR
   ========         ========        ========
Internal State: []

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     -> 1               0: 0     ; loop counter
   EQ_NUM    0 8    2
   JMP_F     5      3
   POP              4
   JMP       9      5
-> POP_UINT8        6
   APPEND_ARR       7
   INC_NUM   0      8
   JMP       1
   END_ARR
   ========         ========        ========
Internal State: []

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     -> 2               0: 0     ; loop counter
   EQ_NUM    0 8    3
   JMP_F     5      4
   POP              5
   JMP       10     6
   POP_UINT8        7
-> APPEND_ARR       8
   INC_NUM   0
   JMP       1
   END_ARR
   ========         ========        ========
Internal State: [1]

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     -> 2               0: 0     ; loop counter
   EQ_NUM    0 8    3
   JMP_F     5      4
   POP              5
   JMP       10     6
   POP_UINT8        7
   APPEND_ARR       8
-> INC_NUM   0
   JMP       1
   END_ARR
   ========         ========        ========
Internal State: [1]

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     -> 2               0: 1     ; loop counter
   EQ_NUM    0 8    3
   JMP_F     5      4
   POP              5
   JMP       10     6
   POP_UINT8        7
   APPEND_ARR       8
   INC_NUM   0
-> JMP       1
   END_ARR
   ========         ========        ========
Internal State: [1]

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     -> 2               0: 1     ; loop counter
-> EQ_NUM    0 8    3
   JMP_F     5      4
   POP              5
   JMP       10     6
   POP_UINT8        7
   APPEND_ARR       8
   INC_NUM   0
   JMP       1
   END_ARR
   ========         ========        ========
Internal State: [1]

...

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     ->                 0: 8     ; loop counter
-> EQ_NUM    0 8
   JMP_F     5
   POP
   JMP       10
   POP_UINT8
   APPEND_ARR
   INC_NUM   0
   JMP       1
   END_ARR
   ========         ========        ========
Internal State: [1, 2, 3, 4, 5, 6, 7, 8]

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     -> true            0: 8     ; loop counter
   EQ_NUM    0 8
-> JMP_F     5
   POP
   JMP       10
   POP_UINT8
   APPEND_ARR
   INC_NUM   0
   JMP       1
   END_ARR
   ========         ========        ========
Internal State: [1, 2, 3, 4, 5, 6, 7, 8]

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     -> true            0: 8     ; loop counter
   EQ_NUM    0 8
   JMP_F     5
-> POP
   JMP       10
   POP_UINT8
   APPEND_ARR
   INC_NUM   0
   JMP       1
   END_ARR
   ========         ========        ========
Internal State: [1, 2, 3, 4, 5, 6, 7, 8]

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     -> true            0: 8     ; loop counter
   EQ_NUM    0 8
   JMP_F     5
-> POP
   JMP       10
   POP_UINT8
   APPEND_ARR
   INC_NUM   0
   JMP       1
   END_ARR
   ========         ========        ========
Internal State: [1, 2, 3, 4, 5, 6, 7, 8]

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     ->                 0: 8     ; loop counter
   EQ_NUM    0 8
   JMP_F     5
   POP
-> JMP       10
   POP_UINT8
   APPEND_ARR
   INC_NUM   0
   JMP       1
   END_ARR
   ========         ========        ========
Internal State: [1, 2, 3, 4, 5, 6, 7, 8]

   INSTR            STACK           DATA
   ========         ========        ========
   START_ARR     ->                 0: 8     ; loop counter
   EQ_NUM    0 8
   JMP_F     5
   POP
   JMP       10
   POP_UINT8
   APPEND_ARR
   INC_NUM   0
   JMP       1
-> END_ARR
   ========         ========        ========
Internal State: [1, 2, 3, 4, 5, 6, 7, 8]
```

### Operations

`LOAD a`: Loads value at address `a` onto the stack
`POP_UINT8`: Loads value at address `a` onto the stack
`EQ_NUM a x`: Load number at address `a` and compare to constant `x`. Push the result onto the stack.
`INC_NUM a`: Assumes that address `a` points to a number on the heap. If
`JMP a`: Unconditionally jumps to address `a`.
`JMP_F a`: Pops a value from the stack. If it is false, jumps to address `a`.
`START_ARR`: Pushes and empty array to the stack.
`APPEND_ARR`: Pops the top two values from the stack. The first may be any element, and the second must be an array. Appends the element to the array and pushes the new array back on the stack.

### TODO

- Add encode/decode for custom types via constructor functions that take a set number of arguments.
- Use conditional jumps to take advantage of decoding repeated structure.
-
