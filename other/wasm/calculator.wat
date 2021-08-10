(module
  (func $add (param $lhs i32) (param $rhs i32) (result i32)
    (i32.add
      (get_local $lhs)
      (get_local $rhs)))

  (func $sub (param $lhs i32) (param $rhs i32) (result i32)
    (i32.sub
      (get_local $lhs)
      (get_local $rhs)))

  (func $mul (param $lhs i32) (param $rhs i32) (result i32)
    (i32.mul
      (get_local $lhs)
      (get_local $rhs)))

  (func $div (param $lhs i32) (param $rhs i32) (result i32)
    (i32.div_s
      (get_local $lhs)
      (get_local $rhs)))

  (export "add" (func $add))
  (export "sub" (func $sub))
  (export "mul" (func $mul))
  (export "div" (func $div))
  )
