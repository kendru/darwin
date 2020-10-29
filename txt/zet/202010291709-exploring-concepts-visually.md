---
tags: thinking visualization
created: Thu Oct 29 17:09:14 MDT 2020
---

# Exploring Concepts Visually

Two things happened to me today that made me think about the value of being able to represent
a concept visually in aiding the formation of a good mental model.

The first was that I was struggling to figure out just the right syntax to represent a
particular nested graph query. There were several subtle concepts relating to how references
are traversed and how certain "magic properties" that are not in the database are made
available for querying, both of which added enough complexity that I had a difficult
time trying to anticipate the effect of my changes. I was able to get it working through
trial and error, but I did not find that process especially enlightening. It would have
been much more helpful to see a visual representation of the query plan that the database
would execute. This could also be presented with lower latency, which would further connect
the cause (change in query) to effect (change in plan). There is no substitute for having
a good mental model, which is the only way that a process can feel intuituve, but visual
exploration can help form that mental model.

A poorly executed or misleading visualization is probably always worse than none at all.

The second occasion that I had to think about visualization was talking with a co-worker
about the complexity of authorization systems. He commented that it would be incredibly
valuable to see in real-time how a proposed edit to a policy would affect particular
users and groups of users. There are sometimes so many inter-dependencies between
components of a system that it is almost intractible to understand statically. Simulation
accompanied by a low-latency visual helps to both discover the effect of an action and
develop an intuition that may not come easily through description alone.
