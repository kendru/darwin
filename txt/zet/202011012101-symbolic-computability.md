---
tags: encoding symbolic-computation
created: Sun Nov  1 21:01:43 MST 2020
---

# Symbolic Computability

Numeric problems can be computed purely symbolically, provided a suitable encoding is used. I was recently reading The Reasoned Schemer,[][bib:friedman2005], which teaches relational programming by example. All of the examples are symbolic, yet there are a couple of chapters on math. In these chapters, numeric problems are solved using unification of lists and symbols.

<!-- TODO: Look up symbols used for numerals and give examples. -->

This caused me to recall the use of Church encoding in the lambda calculus. Given Church numerals, it is possible to come up with operators that perform standard math operations and return the correct result as more Church numerals.

<!-- TODO: cite Pierce here --> Additionally, when we program a digital device to do math, it must use some encoding that represents numbers as a binary signal. Our standard bast 10 system with latin numerals is an equally arbitrary encoding, although we have found it convenient.
