#include <physics/collision.h>
#include <cstdio>
int main() {
    physics::AABB a{mathcore::Vec3(0,0,0), mathcore::Vec3(2,2,2)};
    physics::AABB b{mathcore::Vec3(1,1,1), mathcore::Vec3(3,3,3)};
    physics::AABB c{mathcore::Vec3(5,5,5), mathcore::Vec3(6,6,6)};
    if (!a.intersects(b)) { printf("FAIL: should intersect\n"); return 1; }
    if (a.intersects(c)) { printf("FAIL: should not intersect\n"); return 1; }
    printf("PASS: collision\n");
    return 0;
}
