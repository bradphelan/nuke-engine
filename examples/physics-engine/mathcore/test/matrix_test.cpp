#include <mathcore/matrix.h>
#include <cstdio>
int main() {
    mathcore::Mat4 m = mathcore::Mat4::translate(mathcore::Vec3(1,2,3));
    mathcore::Vec3 p = m.transform_point(mathcore::Vec3(0,0,0));
    if (p.x < 0.99f || p.x > 1.01f) { printf("FAIL: translate\n"); return 1; }
    printf("PASS: matrix\n");
    return 0;
}
