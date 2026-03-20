#include <physics/rigid_body.h>
#include <cstdio>
int main() {
    physics::RigidBody body(1.0f);
    body.apply_force(mathcore::Vec3(10, 0, 0));
    body.step(1.0f);
    if (body.position.x < 9.99f || body.position.x > 10.01f) {
        printf("FAIL: rigid_body\n"); return 1;
    }
    printf("PASS: rigid_body\n");
    return 0;
}
