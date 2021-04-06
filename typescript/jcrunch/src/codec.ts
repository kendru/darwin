import vm from "./vm";

export const encode = (object: any): ArrayBuffer => {
  const v = vm({});
  v.writeObj(object);
  return v.state();
};

export const decode = (data: ArrayBuffer): unknown => {
  const v = vm({
    initialState: data,
  });

  v.rewind();

  return v.read();
};
