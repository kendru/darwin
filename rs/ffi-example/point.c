#include <math.h>
#include "point.h"

// point_distance is designed to be called from Rust.
int point_distance(Point* p1, Point* p2) {
  int dx, dy;
  dx = p1->x - p2->x;
  dy = p1->y - p2->y;
  return sqrt(dx*dx + dy*dy);
}
