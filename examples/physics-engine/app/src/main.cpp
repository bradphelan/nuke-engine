#include <mathcore/vec3.h>
#include <simulation/simulation.h>
#include <cstdio>

int main() {
    printf("=== Physics Engine Demo ===\n");
    mathcore::Vec3 gravity(0, -9.81f, 0);
    printf("Gravity: (%.2f, %.2f, %.2f)\n", gravity.x, gravity.y, gravity.z);

    simulation::Simulation sim;
    sim.add_body(1.0f, 0, 10, 0);
    sim.add_body(2.0f, 5, 20, 0);

    for (int i = 0; i < 5; i++) {
        sim.step(0.016f);
        sim.render();
    }
    printf("Bodies: %d\n", sim.body_count());
    printf("=== Done ===\n");
    return 0;
}
