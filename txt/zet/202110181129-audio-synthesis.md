---
tags: ["dsp","synthesis"]
created: Mon Oct 18 11:29:41 MDT 2021
---

# Audio Synthesis

To generate a sine wave with a period of `M` samples, we can use the following
function: `x[n] = sin(2πn/M)`.

Given some sample rate, e.g. 48Khz, there will be 48,000 samples per cycle, so a
1Hz sine wave could be generated with the following: `x[n] = sin(2πn/48000)` or
`x[n] = sin(πn/24000)`. To generate a sine wave of any higher frequency up to
1/2 the sample rate, only the constant M needs to be modified. For example,
generating a 440Hz sine wave can be done by dividing the sample rate of 48,000
by the target frequency, 440 to obtain M of about 109.

```
S    = 48000
F    = 440
M    = S/F
     ≈ 109.09

x[n] = sin(2πn/M)
     ≈ sin(2πn/109.09)
     ≈ sin(0.057596n)
```
