interface Callable<TParams extends unknown[], TAnswer> {
  call(...args: TParams): TAnswer;
}

const thunk = <TParams extends unknown[], TAnswer>(
  fn: (...args: TParams) => TAnswer
): Callable<TParams, TAnswer> => ({
  call: (...args: TParams) => fn(...args),
});

interface Aggregator<TAcc, TElem, TAnswer = TAcc> {
  initial: Callable<[], TAcc>;
  reducer: Callable<[TAcc, TElem], TAcc>;
  finalizer: Callable<[TAcc], TAnswer>;
}

const mapAggregator = <TAccum, TElemOut, TElemIn, TAnswer>(
  a: Aggregator<TAccum, TElemOut, TAnswer>,
  m: Callable<[TElemIn], TElemOut>
): Aggregator<TAccum, TElemIn, TAnswer> => ({
  initial: a.initial,
  reducer: thunk((acc: TAccum, elem: TElemIn) =>
    a.reducer.call(acc, m.call(elem))
  ),
  finalizer: a.finalizer,
});

export const sum: Aggregator<number, number> = {
  initial: thunk(() => 0),
  reducer: thunk((acc, elem: number) => acc + elem),
  finalizer: thunk((x) => x),
};

export const namedNumberSum = mapAggregator(
  sum,
  thunk((elem: "one" | "two" | "three") => {
    switch (elem) {
      case "one":
        return 1;
      case "two":
        return 2;
      case "three":
        return 3;
    }
  })
);

export const aggregate = <TElem, TAnswer>(
  agg: Aggregator<unknown, TElem, TAnswer>,
  coll: TElem[]
): TAnswer => {
  let acc = agg.initial.call();
  for (const elem of coll) {
    acc = agg.reducer.call(acc, elem);
  }
  return agg.finalizer.call(acc);
};
