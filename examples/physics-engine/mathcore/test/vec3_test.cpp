#include <mathcore/vec3.h>
#include <cstdio>
int main() {
    mathcore::Vec3 a(1,0,0), b(0,1,0);
    mathcore::Vec3 c = a.cross(b);
    if (c.z < 0.99f || c.z > 1.01f) { printf("FAIL: cross\n"); return 1; }
    printf("PASS: vec3\n");
    return 0;
}
